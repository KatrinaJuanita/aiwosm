package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// DictDataDao 字典数据数据访问层 对应Java后端的SysDictDataMapper
type DictDataDao struct {
	db *gorm.DB
}

// NewDictDataDao 创建字典数据数据访问层实例
func NewDictDataDao() *DictDataDao {
	return &DictDataDao{
		db: database.GetDB(),
	}
}

// SelectDictDataList 根据条件分页查询字典数据 对应Java后端的selectDictDataList
func (d *DictDataDao) SelectDictDataList(dictData *model.SysDictData) ([]model.SysDictData, error) {
	var dictDatas []model.SysDictData
	query := d.db.Model(&model.SysDictData{})

	// 构建查询条件
	if dictData.DictType != "" {
		query = query.Where("dict_type = ?", dictData.DictType)
	}
	if dictData.DictLabel != "" {
		query = query.Where("dict_label LIKE ?", "%"+dictData.DictLabel+"%")
	}
	if dictData.Status != "" {
		query = query.Where("status = ?", dictData.Status)
	}

	err := query.Order("dict_sort").Find(&dictDatas).Error
	if err != nil {
		fmt.Printf("SelectDictDataList: 查询字典数据列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDictDataList: 查询到字典数据数量=%d\n", len(dictDatas))
	return dictDatas, nil
}

// SelectDictDataListWithPage 分页查询字典数据 对应Java后端的分页查询
func (d *DictDataDao) SelectDictDataListWithPage(dictData *model.SysDictData, pageNum, pageSize int) ([]model.SysDictData, int64, error) {
	fmt.Printf("DictDataDao.SelectDictDataListWithPage: 分页查询字典数据列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)

	var dictDatas []model.SysDictData
	var total int64

	query := d.db.Model(&model.SysDictData{})

	// 构建查询条件
	if dictData.DictType != "" {
		query = query.Where("dict_type = ?", dictData.DictType)
	}
	if dictData.DictLabel != "" {
		query = query.Where("dict_label LIKE ?", "%"+dictData.DictLabel+"%")
	}
	if dictData.Status != "" {
		query = query.Where("status = ?", dictData.Status)
	}

	// 先查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询字典数据总数失败: %v", err)
	}

	// 分页查询数据
	offset := (pageNum - 1) * pageSize
	err = query.Order("dict_sort").Offset(offset).Limit(pageSize).Find(&dictDatas).Error
	if err != nil {
		return nil, 0, fmt.Errorf("分页查询字典数据列表失败: %v", err)
	}

	fmt.Printf("DictDataDao.SelectDictDataListWithPage: 查询到字典数据数量=%d, 总数=%d\n", len(dictDatas), total)
	return dictDatas, total, nil
}

// SelectDictDataByType 根据字典类型查询字典数据 对应Java后端的selectDictDataByType
func (d *DictDataDao) SelectDictDataByType(dictType string) ([]model.SysDictData, error) {
	var dictDatas []model.SysDictData
	err := d.db.Where("status = '0' AND dict_type = ?", dictType).
		Order("dict_sort").Find(&dictDatas).Error
	if err != nil {
		fmt.Printf("SelectDictDataByType: 查询字典数据失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDictDataByType: 查询到字典数据数量=%d, DictType=%s\n", len(dictDatas), dictType)
	return dictDatas, nil
}

// SelectDictDataById 根据字典数据ID查询信息 对应Java后端的selectDictDataById
func (d *DictDataDao) SelectDictDataById(dictCode int64) (*model.SysDictData, error) {
	var dictData model.SysDictData
	err := d.db.Where("dict_code = ?", dictCode).First(&dictData).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectDictDataById: 查询字典数据详情失败: %v\n", err)
		return nil, err
	}

	return &dictData, nil
}

// SelectDictLabel 根据字典类型和字典键值查询字典标签 对应Java后端的selectDictLabel
func (d *DictDataDao) SelectDictLabel(dictType, dictValue string) (string, error) {
	var dictLabel string
	err := d.db.Model(&model.SysDictData{}).
		Select("dict_label").
		Where("dict_type = ? AND dict_value = ? AND status = '0'", dictType, dictValue).
		Scan(&dictLabel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		fmt.Printf("SelectDictLabel: 查询字典标签失败: %v\n", err)
		return "", err
	}

	return dictLabel, nil
}

// CountDictDataByType 查询字典数据 对应Java后端的countDictDataByType
func (d *DictDataDao) CountDictDataByType(dictType string) (int64, error) {
	var count int64
	err := d.db.Model(&model.SysDictData{}).Where("dict_type = ?", dictType).Count(&count).Error
	if err != nil {
		fmt.Printf("CountDictDataByType: 查询字典数据数量失败: %v\n", err)
		return 0, err
	}

	return count, nil
}

// InsertDictData 新增字典数据信息 对应Java后端的insertDictData
func (d *DictDataDao) InsertDictData(dictData *model.SysDictData) error {
	err := d.db.Create(dictData).Error
	if err != nil {
		fmt.Printf("InsertDictData: 新增字典数据失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertDictData: 新增字典数据成功, DictCode=%d\n", dictData.DictCode)
	return nil
}

// UpdateDictData 修改字典数据信息 对应Java后端的updateDictData
func (d *DictDataDao) UpdateDictData(dictData *model.SysDictData) error {
	err := d.db.Where("dict_code = ?", dictData.DictCode).Updates(dictData).Error
	if err != nil {
		fmt.Printf("UpdateDictData: 修改字典数据失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDictData: 修改字典数据成功, DictCode=%d\n", dictData.DictCode)
	return nil
}

// UpdateDictDataType 修改字典数据类型 对应Java后端的updateDictDataType
func (d *DictDataDao) UpdateDictDataType(oldDictType, newDictType string) error {
	err := d.db.Model(&model.SysDictData{}).
		Where("dict_type = ?", oldDictType).
		Update("dict_type", newDictType).Error
	if err != nil {
		fmt.Printf("UpdateDictDataType: 修改字典数据类型失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDictDataType: 修改字典数据类型成功, %s -> %s\n", oldDictType, newDictType)
	return nil
}

// DeleteDictDataById 删除字典数据信息 对应Java后端的deleteDictDataById
func (d *DictDataDao) DeleteDictDataById(dictCode int64) error {
	err := d.db.Where("dict_code = ?", dictCode).Delete(&model.SysDictData{}).Error
	if err != nil {
		fmt.Printf("DeleteDictDataById: 删除字典数据失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteDictDataById: 删除字典数据成功, DictCode=%d\n", dictCode)
	return nil
}

// DeleteDictDataByIds 批量删除字典数据信息 对应Java后端的deleteDictDataByIds
func (d *DictDataDao) DeleteDictDataByIds(dictCodes []int64) error {
	err := d.db.Where("dict_code IN ?", dictCodes).Delete(&model.SysDictData{}).Error
	if err != nil {
		fmt.Printf("DeleteDictDataByIds: 批量删除字典数据失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteDictDataByIds: 批量删除字典数据成功, 数量=%d\n", len(dictCodes))
	return nil
}
