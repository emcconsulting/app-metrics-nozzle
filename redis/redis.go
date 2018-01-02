package redis

import (
	"fmt"
	"log"
	"os"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	redisAddress = kingpin.Flag("redis-address", "Redis host address").Default("").OverrideDefaultFromEnvar("REDIS_HOST").String()
	redisPort = kingpin.Flag("redis-port", "Redis host address").Default("6379").OverrideDefaultFromEnvar("REDIS_PORT").String()
	redisPassword = kingpin.Flag("redis-password", "Password to the Redis server").Default("").OverrideDefaultFromEnvar("REDIS_PASSWORD").String()
	maxConnections = kingpin.Flag("max-connections", "Max connections to Redis").Default("10").OverrideDefaultFromEnvar("REDIS_MAX_CONN").Int()
)

var logger = log.New(os.Stdout, "", 0)
var redisPool *redis.Pool

func Init() {
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		if *redisPassword == ""{
			c, err := redis.Dial("tcp", *redisAddress + ":" + *redisPort)
			if err != nil {
				logger.Println(err)
				return nil, err
			}
			return c, err
		} else{
			c, err := redis.Dial("tcp", *redisAddress + ":" + *redisPort, redis.DialPassword(*redisPassword))
			if err != nil {
				logger.Println(err)
				return nil, err
			}
			return c, err
		}

	}, *maxConnections)

	logger.Println("Initialized redis pool.")
}

func Set(key string, value int64) bool{
	c := redisPool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, value)

	if err != nil {
		logger.Println(fmt.Sprintf("Could not SET %s:%s", key, value), err)
		return false
	} else {
		return true
	}
}

func Get(key string) int64{
	c := redisPool.Get()
	defer c.Close()

	value, err := redis.Int64(c.Do("GET", key))

	if err != nil {
		return 0
	} else {
		return value
	}
}

func Close(){
	defer redisPool.Close()
}