package backupstore

import (
	"fmt"

	"github.com/yasker/backupstore/util"
)

type VolumeInfo struct {
	Name           string
	Driver         string
	Size           int64 `json:",string"`
	Created        string
	LastBackupName string
	Backups        map[string]*BackupInfo `json:",omitempty"`
}

type BackupInfo struct {
	Name            string
	URL             string
	SnapshotName    string
	SnapshotCreated string
	Created         string
	Size            int64 `json:",string"`

	VolumeName    string `json:",omitempty"`
	VolumeDriver  string `json:",omitempty"`
	VolumeSize    int64  `json:",string,omitempty"`
	VolumeCreated string `json:",omitempty"`
}

func addListVolume(volumeName string, driver BackupStoreDriver, storageDriverName string) (*VolumeInfo, error) {
	if volumeName == "" {
		return nil, fmt.Errorf("Invalid empty volume Name")
	}

	if !util.ValidateName(volumeName) {
		return nil, fmt.Errorf("Invalid volume name %v", volumeName)
	}

	backupNames, err := getBackupNamesForVolume(volumeName, driver)
	if err != nil {
		return nil, err
	}

	volume, err := loadVolume(volumeName, driver)
	if err != nil {
		return nil, err
	}
	//Skip any volumes not owned by specified storage driver
	if volume.Driver != storageDriverName {
		return nil, fmt.Errorf("Incompatiable driver: %v with %v", volume.Driver, storageDriverName)
	}

	volumeInfo := fillVolumeInfo(volume)
	for _, backupName := range backupNames {
		backup, err := loadBackup(backupName, volumeName, driver)
		if err != nil {
			return nil, err
		}
		r := fillBackupInfo(backup, driver.GetURL())
		volumeInfo.Backups[r.URL] = r
	}
	return volumeInfo, nil
}

func List(volumeName, destURL, storageDriverName string) (map[string]*VolumeInfo, error) {
	driver, err := GetBackupStoreDriver(destURL)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]*VolumeInfo)
	if volumeName != "" {
		volumeInfo, err := addListVolume(volumeName, driver, storageDriverName)
		if err != nil {
			return nil, err
		}
		resp[volumeName] = volumeInfo
	} else {
		volumeNames, err := getVolumeNames(driver)
		if err != nil {
			return nil, err
		}
		for _, volumeName := range volumeNames {
			volumeInfo, err := addListVolume(volumeName, driver, storageDriverName)
			if err != nil {
				return nil, err
			}
			resp[volumeName] = volumeInfo
		}
	}
	return resp, nil
}

func fillVolumeInfo(volume *Volume) *VolumeInfo {
	return &VolumeInfo{
		Name:           volume.Name,
		Driver:         volume.Driver,
		Size:           volume.Size,
		Created:        volume.CreatedTime,
		LastBackupName: volume.LastBackupName,
		Backups:        make(map[string]*BackupInfo),
	}
}

func fillBackupInfo(backup *Backup, destURL string) *BackupInfo {
	return &BackupInfo{
		Name:            backup.Name,
		URL:             encodeBackupURL(backup.Name, backup.VolumeName, destURL),
		SnapshotName:    backup.SnapshotName,
		SnapshotCreated: backup.SnapshotCreatedAt,
		Created:         backup.CreatedTime,
		Size:            backup.Size,
	}
}

func fillFullBackupInfo(backup *Backup, volume *Volume, destURL string) *BackupInfo {
	info := fillBackupInfo(backup, destURL)
	info.VolumeName = volume.Name
	info.VolumeDriver = volume.Driver
	info.VolumeSize = volume.Size
	info.VolumeCreated = volume.CreatedTime
	return info
}

func InspectBackup(backupURL string) (*BackupInfo, error) {
	driver, err := GetBackupStoreDriver(backupURL)
	if err != nil {
		return nil, err
	}
	backupName, volumeName, err := decodeBackupURL(backupURL)
	if err != nil {
		return nil, err
	}

	volume, err := loadVolume(volumeName, driver)
	if err != nil {
		return nil, err
	}

	backup, err := loadBackup(backupName, volumeName, driver)
	if err != nil {
		return nil, err
	}
	return fillFullBackupInfo(backup, volume, driver.GetURL()), nil
}