package config

import "time"

type MongoConfig struct {
	URI                    string
	Database               string
	ConnectTimeout         time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	HeartbeatInterval      time.Duration
	ServerSelectionTimeout time.Duration
}

func MongoCfg() MongoConfig {
	return MongoConfig{
		URI:                    getString("MONGO_URI", "mongodb://admin:password@mongodb:27017"),
		Database:               getString("MONGO_DB", "air_social_chat"),
		ConnectTimeout:         getDuration("MONGO_CONNECT_TIMEOUT", 10*time.Second),
		MaxPoolSize:            uint64(getInt("MONGO_MAX_POOL", 100)),
		MinPoolSize:            uint64(getInt("MONGO_MIN_POOL", 5)),
		MaxConnIdleTime:        getDuration("MONGO_MAX_IDLE_TIME", time.Minute*10),
		HeartbeatInterval:      getDuration("MONGO_HEARTBEAT_INTERVAL", 5*time.Second),
		ServerSelectionTimeout: getDuration("MONGO_SERVER_SELECTION_TIMEOUT", 5*time.Second),
	}
}
