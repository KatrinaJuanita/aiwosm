package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// DictTypeDao 字典类型数据访问层 对应Java后端的SysDictTypeMapper
type DictTypeDao struct {
	db *gorm.DB
}

// NewDictTypeDao 创建字典类型数据访问层实例
func NewDictTypeDao() *DictTypeDao {
	return &DictTypeDao{
		db: database.GetDB(),
	}
}

// SelectDictTypeList 查询字典类型 对应Java后端的selectDictTypeList
func (d *DictTypeDao) SelectDictTypeList(dictType *model.SysDictType) ([]model.SysDictType, error) {
	var dictTypes []model.SysDictType
	query := d.db.Model(&model.SysDictType{})

	// 构建查询条件
	if dictType.DictName != "" {
		query = query.Where("dict_name LIKE ?", "%"+dictType.DictName+"%")
	}
	if dictType.Status != "" {
		query = query.Where("status = ?", dictType.Status)
	}
	if dictType.DictType != "" {
		query = query.Where("dict_type LIKE ?", "%"+dictType.DictType+"%")
	}

	err := query.Order("dict_id").Find(&dictTypes).Error
	if err != nil {
		fmt.Printf("SelectDictTypeList: 查询字典类型列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDictTypeList: 查询到字典类型数量=%d\n", len(dictTypes))
	return dictTypes, nil
}

// SelectDictTypeListWithPage 分页查询字典类型 对应Java后端的分页查询
func (d *DictTypeDao) SelectDictTypeListWithPage(dictType *model.SysDictType, pageNum, pageSize int) ([]model.SysDictType, int64, error) {
	fmt.Printf("DictTypeDao.SelectDictTypeListWithPage: 分页查询字典类型列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)

	var dictTypes []model.SysDictType
	var total int64

	query := d.db.Model(&model.SysDictType{})

	// 构建查询条件
	if dictType.DictName != "" {
		query = query.Where("dict_name LIKE ?", "%"+dictType.DictName+"%")
	}
	if dictType.Status != "" {
		query = query.Where("status = ?", dictType.Status)
	}
	if dictType.DictType != "" {
		query = query.Where("dict_type LIKE ?", "%"+dictType.DictType+"%")
	}

	// 先查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询字典类型总数失败: %v", err)
	}

	// 分页查询数据
	offset := (pageNum - 1) * pageSize
	err = query.Order("dict_id").Offset(offset).Limit(pageSize).Find(&dictTypes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("分页查询字典类型列表失败: %v", err)
	}

	fmt.Printf("DictTypeDao.SelectDictTypeListWithPage: 查询到字典类型数量=%d, 总数=%d\n", len(dictTypes), total)
	return dictTypes, total, nil
}

// SelectDictTypeAll 查询所有字典类型 对应Java后端的selectDictTypeAll
func (d *DictTypeDao) SelectDictTypeAll() ([]model.SysDictType, error) {
	var dictTypes []model.SysDictType
	err := d.db.Order("dict_id").Find(&dictTypes).Error
	if err != nil {
		fmt.Printf("SelectDictTypeAll: 查询所有字典类型失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDictTypeAll: 查询到字典类型数量=%d\n", len(dictTypes))
	return dictTypes, nil
}

// SelectDictTypeById 根据字典类型ID查询信息 对应Java后端的selectDictTypeById
func (d *DictTypeDao) SelectDictTypeById(dictId int64) (*model.SysDictType, error) {
	var dictType model.SysDictType
	err := d.db.Where("dict_id = ?", dictId).First(&dictType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectDictTypeById: 查询字典类型详情失败: %v\n", err)
		return nil, err
	}

	return &dictType, nil
}

// SelectDictTypeByType 根据字典类型查询信息 对应Java后端的selectDictTypeByType
func (d *DictTypeDao) SelectDictTypeByType(dictType string) (*model.SysDictType, error) {
	var dict model.SysDictType
	err := d.db.Where("dict_type = ?", dictType).First(&dict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectDictTypeByType: 查询字典类型失败: %v\n", err)
		return nil, err
	}

	return &dict, nil
}

// CheckDictTypeUnique 校验字典类型称是否唯一 对应Java后端的checkDictTypeUnique
func (d *DictTypeDao) CheckDictTypeUnique(dictType *model.SysDictType) (*model.SysDictType, error) {
	var existDict model.SysDictType
	query := d.db.Where("dict_type = ?", dictType.DictType)

	// 如果是修改操作，排除自己
	if dictType.DictID != 0 {
		query = query.Where("dict_id != ?", dictType.DictID)
	}

	err := query.First(&existDict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("CheckDictTypeUnique: 校验字典类型唯一性失败: %v\n", err)
		return nil, err
	}

	return &existDict, nil
}

// InsertDictType 新增字典类型信息 对应Java后端的insertDictType
func (d *DictTypeDao) InsertDictType(dictType *model.SysDictType) error {
	err := d.db.Create(dictType).Error
	if err != nil {
		fmt.Printf("InsertDictType: 新增字典类型失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertDictType: 新增字典类型成功, DictID=%d\n", dictType.DictID)
	return nil
}

// UpdateDictType 修改字典类型信息 对应Java后端的updateDictType
func (d *DictTypeDao) UpdateDictType(dictType *model.SysDictType) error {
	err := d.db.Where("dict_id = ?", dictType.DictID).Updates(dictType).Error
	if err != nil {
		fmt.Printf("UpdateDictType: 修改字典类型失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDictType: 修改字典类型成功, DictID=%d\n", dictType.DictID)
	return nil
}

// DeleteDictTypeById 删除字典类型信息 对应Java后端的deleteDictTypeById
func (d *DictTypeDao) DeleteDictTypeById(dictId int64) error {
	err := d.db.Where("dict_id = ?", dictId).Delete(&model.SysDictType{}).Error
	if err != nil {
		fmt.Printf("DeleteDictTypeById: 删除字典类型失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteDictTypeById: 删除字典类型成功, DictID=%d\n", dictId)
	return nil
}

// DeleteDictTypeByIds 批量删除字典类型信息 对应Java后端的deleteDictTypeByIds
func (d *DictTypeDao) DeleteDictTypeByIds(dictIds []int64) error {
	err := d.db.Where("dict_id IN ?", dictIds).Delete(&model.SysDictType{}).Error
	if err != nil {
		fmt.Printf("DeleteDictTypeByIds: 批量删除字典类型失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteDictTypeByIds: 批量删除字典类型成功, 数量=%d\n", len(dictIds))
	return nil
}
