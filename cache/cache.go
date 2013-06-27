package cache

type ICache interface {
	Init(map[string]string) bool
	Get(string) (interface{}, error)
	GetString(string) (string, error)
	GetInt(string) (int, error)
	GetInt64(string) (int64, error)
	GetFloat64(string) (float64, error)
	Set(string, interface{}, int) error
	Exists(string) (bool, error)
	Delete(string) error
}

var (
	cache ICache
)

func Init(config map[string]string) {
	cache = &RedisCache{}
	if !cache.Init(config) {
		cache = nil
	}

}

func IsEnable() bool {
	return cache != nil
}

func Exists(key string) (bool, error) {
	return cache.Exists(key)
}

func Get(key string) (interface{}, error) {
	return cache.Get(key)
}

func GetString(key string) (string, error) {
	return cache.GetString(key)
}

func GetInt(key string) (int, error) {
	return cache.GetInt(key)
}

func GetInt64(key string) (int64, error) {
	return cache.GetInt64(key)
}

func GetFloat64(key string) (float64, error) {
	return cache.GetFloat64(key)
}

func Set(key string, value interface{}, expire int) error {
	return cache.Set(key, value, expire)
}

func Delete(key string) error {
	return cache.Delete(key)
}
