package person

import (
	"fmt"
	"nft_backend/internal/logger"

	"gorm.io/gorm"
)

// 做一个struct用于管理对象属性, 类似于无方法的class类
type Repository struct {
	db *gorm.DB
}

// 做一个构造函数从外部传入db对象
func NewRepo(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// 下面所有方法都是Repository的方法, 接收者都是*Repository

// 创建记录
func (r *Repository) Create(person *Person) error {
	fmt.Println("创建person记录")
	return r.db.Create(person).Error
}

func (r *Repository) ReadSingle(id int, ownerId *int) (*Person, error) {
	var p Person
	fmt.Println("读取person记录")
	if ownerId == nil {
		if err := r.db.First(&p, id).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Where("user_id = ?", *ownerId).First(&p, id).Error; err != nil {
			return nil, err
		}
	}
	return &p, nil
}

func (r *Repository) ReadList(rows int, ownerId *int) ([]Person, error) {
	var list []Person
	fmt.Println("读取person记录列表")
	if ownerId == nil {
		if err := r.db.Order("id desc").Limit(rows).Find(&list).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Where("user_id = ?", *ownerId).Order("id desc").Limit(rows).Find(&list).Error; err != nil {
			return nil, err
		}
	}

	return list, nil
}

// 分页查询

func (r *Repository) ReadPage(q PersonPageQuery, ownerId *int) (PersonPageEntity, error) {
	var list []Person
	var total int64

	logger.L.Info("读取person记录分页")

	// 1. 构建一个基础查询（后面可以在这里加 Where 条件）
	query := r.db.Model(&Person{})

	// 构造条件查询
	if q.City != "" {
		query = query.Where("city = ?", q.City)
	}
	if q.Sex != "" {
		query = query.Where("sex = ?", q.Sex)
	}
	pageNo := q.PageNo
	pageSize := q.PageSize
	// 2. 再查总数
	if ownerId == nil {
		if err := query.Count(&total).Error; err != nil {
			return PersonPageEntity{}, err
		}
		// 3. 再查当前页数据
		if err := query.
			Offset((pageNo - 1) * pageSize).
			Limit(pageSize).
			Find(&list).Error; err != nil {
			return PersonPageEntity{}, err
		}
	} else {
		query = query.Where("user_id = ?", *ownerId)
		if err := query.Count(&total).Error; err != nil {
			return PersonPageEntity{}, err
		}
		// 3. 再查当前页数据
		if err := query.
			Offset((pageNo - 1) * pageSize).
			Limit(pageSize).
			Find(&list).Error; err != nil {
			return PersonPageEntity{}, err
		}
	}

	// 4. 封装成分页结构返回（带 json tag 的）
	return PersonPageEntity{
		PageNo:   pageNo,
		PageSize: pageSize,
		Total:    int(total),
		List:     list,
	}, nil
}

// 第一种update, 先查后改
func (r *Repository) Update(person *Person, ownerId *int) error {
	fmt.Println("更新person记录")

	db := r.db.Model(&Person{}).Where("id = ?", person.ID)

	if ownerId != nil {
		db = db.Where("user_id = ?", *ownerId)
	}
	return db.Updates(person).Error

}

// 删除
func (r *Repository) Delete(id int, ownerId *int) error {
	fmt.Println("删除person记录")
	if ownerId == nil {
		return r.db.Delete(&Person{}, id).Error
	} else {
		return r.db.Where("user_id = ?", *ownerId).Delete(&Person{}, id).Error
	}

}
