package cache

type ICache interface {
	Init(map[string]string) bool
	Get([]byte) (interface{}, error)
	GetString([]byte) (string, error)
	GetInt([]byte) (int, error)
	GetInt64([]byte) (int64, error)
	GetFloat64([]byte) (float64, error)
	Set([]byte, interface{}, int) error
	Exists([]byte) (bool, error)
	Delete([]byte) error
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

func Exists(bkey []byte) (bool, error) {
	return cache.Exists(bkey)
}

func Get(bkey []byte) (interface{}, error) {
	return cache.Get(bkey)
}

func GetString(bkey []byte) (string, error) {
	return cache.GetString(bkey)
}

func GetInt(bkey []byte) (int, error) {
	return cache.GetInt(bkey)
}

func GetInt64(bkey []byte) (int64, error) {
	return cache.GetInt64(bkey)
}

func GetFloat64(bkey []byte) (float64, error) {
	return cache.GetFloat64(bkey)
}

func Set(bkey []byte, value interface{}, expire int) error {
	return cache.Set(bkey, value, expire)
}

func Delete(bkey []byte) error {
	return cache.Delete(bkey)
}
