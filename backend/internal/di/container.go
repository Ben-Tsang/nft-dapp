package di

import (
	"errors"
	"reflect"
	"sync"
)

// Container 极简DI容器核心结构
type Container struct {
	instances map[reflect.Type]interface{} // 存储实例（键=类型，值=实例）
	mu        sync.RWMutex                 // 线程安全锁
}

// 全局单例容器（包级私有，对外只暴露方法）
var (
	globalContainer *Container
	once            sync.Once // 保证容器只初始化一次
)

// Init 初始化全局DI容器（程序启动时调用一次）
func Init() {
	once.Do(func() {
		globalContainer = &Container{
			instances: make(map[reflect.Type]interface{}),
		}
	})
}

// Register 注册实例到容器（泛型保证类型安全）
// 示例：di.Register(&status.NFTServiceImpl{})
func Register[T any](instance T) error {
	if globalContainer == nil {
		return errors.New("DI容器未初始化，请先调用 di.Init()")
	}

	globalContainer.mu.Lock()
	defer globalContainer.mu.Unlock()

	// 获取实例的底层类型（统一去指针，避免重复注册）
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// 检查是否已注册
	if _, exists := globalContainer.instances[typ]; exists {
		return errors.New("实例已注册：" + typ.Name())
	}

	globalContainer.instances[typ] = instance
	return nil
}

// Resolve 从容器获取实例（泛型保证类型安全）
// 示例：svc, err := di.Resolve[status.NFTService]()
func Resolve[T any]() (T, error) {
	var zero T // 泛型零值，用于返回错误时的默认值
	if globalContainer == nil {
		return zero, errors.New("DI容器未初始化，请先调用 di.Init()")
	}

	globalContainer.mu.RLock()
	defer globalContainer.mu.RUnlock()

	// 获取目标类型（统一去指针）
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// 从容器查找实例
	inst, exists := globalContainer.instances[typ]
	if !exists {
		return zero, errors.New("实例未注册：" + typ.Name())
	}

	// 类型断言（保证返回类型正确）
	result, ok := inst.(T)
	if !ok {
		return zero, errors.New("类型不匹配：" + typ.Name())
	}

	return result, nil
}
