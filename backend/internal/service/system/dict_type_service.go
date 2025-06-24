package system

import (
	"fmt"
	"regexp"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/dict"
)

// DictTypeService 字典类型服务 对应Java后端的ISysDictTypeService
type DictTypeService struct {
	dictTypeDao *dao.DictTypeDao
	dictDataDao *dao.DictDataDao
}

// NewDictTypeService 创建字典类型服务实例
func NewDictTypeService() *DictTypeService {
	return &DictTypeService{
		dictTypeDao: dao.NewDictTypeDao(),
		dictDataDao: dao.NewDictDataDao(),
	}
}

// SelectDictTypeList 根据条件分页查询字典类型 对应Java后端的selectDictTypeList
func (s *DictTypeService) SelectDictTypeList(dictType *model.SysDictType) ([]model.SysDictType, error) {
	fmt.Printf("DictTypeService.SelectDictTypeList: 查询字典类型列表\n")
	return s.dictTypeDao.SelectDictTypeList(dictType)
}

// SelectDictTypeListWithPage 分页查询字典类型 对应Java后端的分页查询
func (s *DictTypeService) SelectDictTypeListWithPage(dictType *model.SysDictType, pageNum, pageSize int) ([]model.SysDictType, int64, error) {
	fmt.Printf("DictTypeService.SelectDictTypeListWithPage: 分页查询字典类型列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)
	return s.dictTypeDao.SelectDictTypeListWithPage(dictType, pageNum, pageSize)
}

// SelectDictTypeAll 查询所有字典类型 对应Java后端的selectDictTypeAll
func (s *DictTypeService) SelectDictTypeAll() ([]model.SysDictType, error) {
	fmt.Printf("DictTypeService.SelectDictTypeAll: 查询所有字典类型\n")
	return s.dictTypeDao.SelectDictTypeAll()
}

// SelectDictTypeById 根据字典类型ID查询信息 对应Java后端的selectDictTypeById
func (s *DictTypeService) SelectDictTypeById(dictId int64) (*model.SysDictType, error) {
	fmt.Printf("DictTypeService.SelectDictTypeById: 查询字典类型详情, DictID=%d\n", dictId)
	return s.dictTypeDao.SelectDictTypeById(dictId)
}

// SelectDictTypeByType 根据字典类型查询信息 对应Java后端的selectDictTypeByType
func (s *DictTypeService) SelectDictTypeByType(dictType string) (*model.SysDictType, error) {
	fmt.Printf("DictTypeService.SelectDictTypeByType: 查询字典类型, DictType=%s\n", dictType)
	return s.dictTypeDao.SelectDictTypeByType(dictType)
}

// ValidateDictType 验证字典类型格式 对应Java后端的@Pattern验证
func (s *DictTypeService) ValidateDictType(dictType string) error {
	// 字典类型必须以字母开头，且只能为（小写字母，数字，下划线）
	pattern := `^[a-z][a-z0-9_]*$`
	matched, err := regexp.MatchString(pattern, dictType)
	if err != nil {
		return fmt.Errorf("正则表达式验证失败: %v", err)
	}
	if !matched {
		return fmt.Errorf("字典类型必须以字母开头，且只能为（小写字母，数字，下划线）")
	}
	return nil
}

// CheckDictTypeUnique 校验字典类型是否唯一 对应Java后端的checkDictTypeUnique
func (s *DictTypeService) CheckDictTypeUnique(dictType *model.SysDictType) bool {
	fmt.Printf("DictTypeService.CheckDictTypeUnique: 校验字典类型唯一性, DictType=%s\n", dictType.DictType)

	existDict, err := s.dictTypeDao.CheckDictTypeUnique(dictType)
	if err != nil {
		return false
	}

	// 如果不存在重复，返回true
	return existDict == nil
}

// InsertDictType 新增保存字典类型信息 对应Java后端的insertDictType
func (s *DictTypeService) InsertDictType(dictType *model.SysDictType) error {
	fmt.Printf("DictTypeService.InsertDictType: 新增字典类型, DictType=%s\n", dictType.DictType)

	// 验证字典类型格式 对应Java后端的@Pattern验证
	if err := s.ValidateDictType(dictType.DictType); err != nil {
		return err
	}

	// 设置创建时间
	now := time.Now()
	dictType.CreateTime = &now

	return s.dictTypeDao.InsertDictType(dictType)
}

// UpdateDictType 修改保存字典类型信息 对应Java后端的updateDictType
func (s *DictTypeService) UpdateDictType(dictType *model.SysDictType) error {
	fmt.Printf("DictTypeService.UpdateDictType: 修改字典类型, DictID=%d\n", dictType.DictID)

	// 验证字典类型格式 对应Java后端的@Pattern验证
	if err := s.ValidateDictType(dictType.DictType); err != nil {
		return err
	}

	// 查询旧的字典类型
	oldDictType, err := s.dictTypeDao.SelectDictTypeById(dictType.DictID)
	if err != nil {
		return err
	}

	// 设置更新时间
	now := time.Now()
	dictType.UpdateTime = &now

	// 修改字典类型
	err = s.dictTypeDao.UpdateDictType(dictType)
	if err != nil {
		return err
	}

	// 如果字典类型发生变化，需要更新字典数据中的类型
	if oldDictType != nil && oldDictType.DictType != dictType.DictType {
		err = s.dictDataDao.UpdateDictDataType(oldDictType.DictType, dictType.DictType)
		if err != nil {
			return err
		}

		// 删除旧的缓存
		dict.RemoveDictCache(oldDictType.DictType)
	}

	// 删除缓存
	dict.RemoveDictCache(dictType.DictType)

	return nil
}

// DeleteDictTypeByIds 批量删除字典类型信息 对应Java后端的deleteDictTypeByIds
func (s *DictTypeService) DeleteDictTypeByIds(dictIds []int64) error {
	fmt.Printf("DictTypeService.DeleteDictTypeByIds: 批量删除字典类型, DictIDs=%v\n", dictIds)

	for _, dictId := range dictIds {
		// 查询字典类型信息
		dictType, err := s.dictTypeDao.SelectDictTypeById(dictId)
		if err != nil {
			return err
		}

		if dictType == nil {
			continue
		}

		// 检查是否有关联的字典数据 对应Java后端的countDictDataByType检查
		count, err := s.dictDataDao.CountDictDataByType(dictType.DictType)
		if err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("%s已分配,不能删除", dictType.DictName)
		}

		// 删除字典类型
		err = s.dictTypeDao.DeleteDictTypeById(dictId)
		if err != nil {
			return err
		}

		// 删除缓存
		dict.RemoveDictCache(dictType.DictType)
	}

	return nil
}

// ResetDictCache 重置字典缓存数据 对应Java后端的resetDictCache
func (s *DictTypeService) ResetDictCache() error {
	fmt.Printf("DictTypeService.ResetDictCache: 重置字典缓存\n")

	// 清空所有缓存
	dict.ClearDictCache()

	// 查询所有字典类型
	dictTypes, err := s.dictTypeDao.SelectDictTypeAll()
	if err != nil {
		return err
	}

	// 重新加载所有字典数据到缓存
	for _, dictType := range dictTypes {
		dictDatas, err := s.dictDataDao.SelectDictDataByType(dictType.DictType)
		if err == nil {
			dict.SetDictCache(dictType.DictType, dictDatas)
		}
	}

	fmt.Printf("DictTypeService.ResetDictCache: 重置字典缓存完成, 字典类型数量=%d\n", len(dictTypes))
	return nil
}

// LoadingDictCache 加载字典缓存数据 对应Java后端的loadingDictCache
func (s *DictTypeService) LoadingDictCache() {
	fmt.Printf("DictTypeService.LoadingDictCache: 加载字典缓存\n")

	// 异步加载缓存，避免阻塞启动
	go func() {
		err := s.ResetDictCache()
		if err != nil {
			fmt.Printf("DictTypeService.LoadingDictCache: 加载字典缓存失败: %v\n", err)
		} else {
			fmt.Printf("DictTypeService.LoadingDictCache: 加载字典缓存成功\n")
		}
	}()
}
