package system

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/dict"
)

// DictDataService 字典数据服务 对应Java后端的ISysDictDataService
type DictDataService struct {
	dictDataDao *dao.DictDataDao
}

// NewDictDataService 创建字典数据服务实例
func NewDictDataService() *DictDataService {
	return &DictDataService{
		dictDataDao: dao.NewDictDataDao(),
	}
}

// SelectDictDataList 根据条件分页查询字典数据 对应Java后端的selectDictDataList
func (s *DictDataService) SelectDictDataList(dictData *model.SysDictData) ([]model.SysDictData, error) {
	fmt.Printf("DictDataService.SelectDictDataList: 查询字典数据列表\n")
	return s.dictDataDao.SelectDictDataList(dictData)
}

// SelectDictDataListWithPage 分页查询字典数据 对应Java后端的分页查询
func (s *DictDataService) SelectDictDataListWithPage(dictData *model.SysDictData, pageNum, pageSize int) ([]model.SysDictData, int64, error) {
	fmt.Printf("DictDataService.SelectDictDataListWithPage: 分页查询字典数据列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)
	return s.dictDataDao.SelectDictDataListWithPage(dictData, pageNum, pageSize)
}

// SelectDictDataByType 根据字典类型查询字典数据 对应Java后端的selectDictDataByType
func (s *DictDataService) SelectDictDataByType(dictType string) ([]model.SysDictData, error) {
	fmt.Printf("DictDataService.SelectDictDataByType: 查询字典数据, DictType=%s\n", dictType)

	// 先从缓存获取
	dictDatas := dict.GetDictCache(dictType)
	if dictDatas != nil {
		return dictDatas, nil
	}

	// 缓存未命中，从数据库查询
	dictDatas, err := s.dictDataDao.SelectDictDataByType(dictType)
	if err != nil {
		return nil, err
	}

	// 设置缓存
	dict.SetDictCache(dictType, dictDatas)

	return dictDatas, nil
}

// SelectDictDataById 根据字典数据ID查询信息 对应Java后端的selectDictDataById
func (s *DictDataService) SelectDictDataById(dictCode int64) (*model.SysDictData, error) {
	fmt.Printf("DictDataService.SelectDictDataById: 查询字典数据详情, DictCode=%d\n", dictCode)
	return s.dictDataDao.SelectDictDataById(dictCode)
}

// SelectDictLabel 根据字典类型和字典键值查询字典标签 对应Java后端的selectDictLabel
func (s *DictDataService) SelectDictLabel(dictType, dictValue string) (string, error) {
	fmt.Printf("DictDataService.SelectDictLabel: 查询字典标签, DictType=%s, DictValue=%s\n", dictType, dictValue)
	return s.dictDataDao.SelectDictLabel(dictType, dictValue)
}

// DeleteDictDataByIds 批量删除字典数据信息 对应Java后端的deleteDictDataByIds
func (s *DictDataService) DeleteDictDataByIds(dictCodes []int64) error {
	fmt.Printf("DictDataService.DeleteDictDataByIds: 批量删除字典数据, 数量=%d\n", len(dictCodes))

	for _, dictCode := range dictCodes {
		// 查询字典数据信息
		dictData, err := s.dictDataDao.SelectDictDataById(dictCode)
		if err != nil {
			return err
		}

		// 删除字典数据
		err = s.dictDataDao.DeleteDictDataById(dictCode)
		if err != nil {
			return err
		}

		// 更新缓存
		dictDatas, err := s.dictDataDao.SelectDictDataByType(dictData.DictType)
		if err != nil {
			return err
		}
		dict.SetDictCache(dictData.DictType, dictDatas)
	}

	return nil
}

// InsertDictData 新增保存字典数据信息 对应Java后端的insertDictData
func (s *DictDataService) InsertDictData(dictData *model.SysDictData) error {
	fmt.Printf("DictDataService.InsertDictData: 新增字典数据, DictType=%s\n", dictData.DictType)

	// 设置创建时间
	now := time.Now()
	dictData.CreateTime = &now

	// 新增字典数据
	err := s.dictDataDao.InsertDictData(dictData)
	if err != nil {
		return err
	}

	// 删除缓存
	dict.RemoveDictCache(dictData.DictType)

	return nil
}

// UpdateDictData 修改保存字典数据信息 对应Java后端的updateDictData
func (s *DictDataService) UpdateDictData(dictData *model.SysDictData) error {
	fmt.Printf("DictDataService.UpdateDictData: 修改字典数据, DictCode=%d\n", dictData.DictCode)

	// 设置更新时间
	now := time.Now()
	dictData.UpdateTime = &now

	// 修改字典数据
	err := s.dictDataDao.UpdateDictData(dictData)
	if err != nil {
		return err
	}

	// 删除缓存
	dict.RemoveDictCache(dictData.DictType)

	return nil
}

// CountDictDataByType 查询字典数据 对应Java后端的countDictDataByType
func (s *DictDataService) CountDictDataByType(dictType string) (int64, error) {
	fmt.Printf("DictDataService.CountDictDataByType: 查询字典数据数量, DictType=%s\n", dictType)
	return s.dictDataDao.CountDictDataByType(dictType)
}
