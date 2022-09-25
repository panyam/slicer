package db

import (
	"errors"
	"gorm.io/gorm"
	"log"
	"time"
)

type DB struct {
	storage *gorm.DB
}

func New(gormdb *gorm.DB) ControlDB {
	gormdb.AutoMigrate(&Shard{})
	gormdb.AutoMigrate(&Target{})
	gormdb.AutoMigrate(&ShardTarget{})
	return &DB{storage: gormdb}
}

func (ctrldb *DB) GetTargets(addresses ...string) ([]*Target, error) {
	if len(addresses) == 1 {
		var out Target
		result := ctrldb.storage.First(&out, "address = ?", addresses[0])
		err := result.Error
		if err != nil {
			return []*Target{}, err
		} else {
			return []*Target{&out}, err
		}
	} else { // batch get
		var out []*Target
		result := ctrldb.storage.Where("address IN ?", addresses).Find(&out)
		err := result.Error
		if err != nil {
			return []*Target{}, err
		} else {
			return out, err
		}
	}
}

func (ctrldb *DB) SaveTarget(target *Target) (err error) {
	target.UpdatedAt = time.Now()
	db := ctrldb.storage
	q := db.Model(target).Where("address = ? and version = ?", target.Address, target.Version)
	result := q.UpdateColumns(map[string]interface{}{
		"status":    target.Status,
		"tags":      target.Tags,
		"pinged_at": target.PingedAt,
		"version":   target.Version + 1,
	})

	err = result.Error
	if err != nil {
		log.Println("SaveTargets RowsAffected: ", result.RowsAffected)
	}
	if err == nil && result.RowsAffected == 0 {
		// Must have failed due to versioning
		// This means no row existed so we can do a create
		result = ctrldb.storage.Model(&Target{}).Create(target)
		err = result.Error
		if err != nil {
			log.Println("SaveTargets Create Error: ", err)
			err = UpdateFailed
		}
	} else if err != nil {
		log.Println("SaveTargets Update Error: ", err, result.RowsAffected)
	}
	return
}

func (ctrldb *DB) DeleteTargets(addresses ...string) error {
	if len(addresses) == 1 {
		return ctrldb.storage.Where("target_address = ?", addresses[0]).Delete(&Target{}).Error
	} else {
		return ctrldb.storage.Where("target_address in (?)", addresses).Delete(&Target{}).Error
	}
}

func (ctrldb *DB) GetShard(key string) (*[]Shard, error) {
	var out *[]Shard
	result := ctrldb.storage.Where("key = ?", key).Find(&out)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
	}
	return out, err
}

func (ctrldb *DB) SaveShard(shard *Shard) (err error) {
	shard.UpdatedAt = time.Now()
	db := ctrldb.storage

	q := db.Model(shard).Where("key = ? and version = ?", shard.Key, shard.Version)
	result := q.UpdateColumns(map[string]interface{}{
		"version": shard.Version + 1,
	})

	err = result.Error
	if err == nil && result.RowsAffected == 0 {
		// Must have failed due to versioning
		result = ctrldb.storage.Model(&Shard{}).Create(shard)
		err = result.Error
		log.Println("Shard Save Error: ", err)
	} else {
		log.Println("Shard Save Error 2: ", err)
	}
	return
}

func (ctrldb *DB) DeleteShard(key string) error {
	return ctrldb.storage.Where("key = ?", key).Delete(&Shard{}).Error
}

func (ctrldb *DB) GetShardTarget(key string) (*[]ShardTarget, error) {
	var out *[]ShardTarget
	result := ctrldb.storage.Where("shard_key = ?", key).Find(&out)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
	}
	return out, err
}

func (ctrldb *DB) SaveShardTarget(shard_target *ShardTarget) (err error) {
	shard_target.UpdatedAt = time.Now()
	db := ctrldb.storage

	q := db.Model(shard_target).Where("key = ? and version = ?", shard_target.ShardKey, shard_target.Version)
	result := q.UpdateColumns(map[string]interface{}{
		"version": shard_target.Version + 1,
	})

	err = result.Error
	if err == nil && result.RowsAffected == 0 {
		// Must have failed due to versioning
		result = ctrldb.storage.Model(&ShardTarget{}).Create(shard_target)
		err = result.Error
		log.Println("ShardTarget Save Error: ", err)
	} else {
		log.Println("ShardTarget Save Error 2: ", err)
	}
	return
}

func (ctrldb *DB) DeleteShardTargets(key string, addresses ...string) error {
	if len(addresses) == 0 {
		return ctrldb.storage.Where("shard_key = ?", key).Delete(&ShardTarget{}).Error
	} else {
		return ctrldb.storage.Where("shard_key = ? AND target_address IN ?", key, addresses).Delete(&ShardTarget{}).Error
	}
}
