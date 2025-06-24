<template>
  <div class="app-container">
    <!-- 查询条件 -->
    <el-form
      :model="queryParams"
      ref="queryForm"
      size="small"
      :inline="true"
      v-show="showSearch"
      label-width="68px"
    >
      <el-form-item label="物料编码" prop="materialCode">
        <el-input
          v-model="queryParams.materialCode"
          placeholder="请输入物料编码"
          clearable
          style="width: 240px"
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item label="物料名称" prop="materialName">
        <el-input
          v-model="queryParams.materialName"
          placeholder="请输入物料名称"
          clearable
          style="width: 240px"
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item label="物料类型" prop="materialType">
        <el-select
          v-model="queryParams.materialType"
          placeholder="请选择物料类型"
          clearable
          style="width: 240px"
        >
          <el-option label="原材料" value="RAW_MATERIAL" />
          <el-option label="半成品" value="SEMI_FINISHED" />
          <el-option label="成品" value="FINISHED_PRODUCT" />
          <el-option label="包装材料" value="PACKAGING" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select
          v-model="queryParams.status"
          placeholder="物料状态"
          clearable
          style="width: 240px"
        >
          <el-option label="正常" value="0" />
          <el-option label="停用" value="1" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery"
          >搜索</el-button
        >
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 工具栏 -->
    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button
          type="primary"
          plain
          icon="el-icon-plus"
          size="mini"
          @click="handleAdd"
          v-hasPermi="['inventory:material:add']"
          >新增</el-button
        >
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="success"
          plain
          icon="el-icon-edit"
          size="mini"
          :disabled="single"
          @click="handleUpdate"
          v-hasPermi="['inventory:material:edit']"
          >修改</el-button
        >
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="danger"
          plain
          icon="el-icon-delete"
          size="mini"
          :disabled="multiple"
          @click="handleDelete"
          v-hasPermi="['inventory:material:remove']"
          >删除</el-button
        >
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="warning"
          plain
          icon="el-icon-download"
          size="mini"
          @click="handleExport"
          v-hasPermi="['inventory:material:export']"
          >导出</el-button
        >
      </el-col>
      <el-col :span="1.5">
        <el-button type="info" plain icon="el-icon-warning" size="mini" @click="handleLowStock"
          >库存预警</el-button
        >
      </el-col>
      <right-toolbar :showSearch.sync="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <!-- 数据表格 -->
    <el-table v-loading="loading" :data="materialList" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column label="物料编码" align="center" prop="materialCode" width="120" />
      <el-table-column
        label="物料名称"
        align="center"
        prop="materialName"
        width="150"
        show-overflow-tooltip
      />
      <el-table-column label="物料类型" align="center" prop="materialType" width="100">
        <template slot-scope="scope">
          <dict-tag :options="dict.type.material_type" :value="scope.row.materialType" />
        </template>
      </el-table-column>
      <el-table-column
        label="规格型号"
        align="center"
        prop="specification"
        width="120"
        show-overflow-tooltip
      />
      <el-table-column label="计量单位" align="center" prop="unit" width="80" />
      <el-table-column label="品牌" align="center" prop="brand" width="100" />
      <el-table-column label="当前库存" align="center" width="100">
        <template slot-scope="scope">
          <span :class="getStockClass(scope.row.stockStatus)">
            {{ scope.row.currentStock }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="安全库存" align="center" prop="safetyStock" width="100" />
      <el-table-column label="标准价格" align="center" width="100">
        <template slot-scope="scope"> ¥{{ scope.row.standardPrice }} </template>
      </el-table-column>
      <el-table-column
        label="供应商"
        align="center"
        prop="supplierName"
        width="120"
        show-overflow-tooltip
      />
      <el-table-column label="状态" align="center" width="80">
        <template slot-scope="scope">
          <dict-tag :options="dict.type.sys_normal_disable" :value="scope.row.status" />
        </template>
      </el-table-column>
      <el-table-column label="创建时间" align="center" prop="createdTime" width="160">
        <template slot-scope="scope">
          <span>{{ parseTime(scope.row.createdTime, '{y}-{m}-{d} {h}:{i}:{s}') }}</span>
        </template>
      </el-table-column>
      <el-table-column
        label="操作"
        align="center"
        width="160"
        class-name="small-padding fixed-width"
      >
        <template slot-scope="scope">
          <el-button
            size="mini"
            type="text"
            icon="el-icon-view"
            @click="handleView(scope.row)"
            v-hasPermi="['inventory:material:query']"
            >详情</el-button
          >
          <el-button
            size="mini"
            type="text"
            icon="el-icon-edit"
            @click="handleUpdate(scope.row)"
            v-hasPermi="['inventory:material:edit']"
            >修改</el-button
          >
          <el-button
            size="mini"
            type="text"
            icon="el-icon-delete"
            @click="handleDelete(scope.row)"
            v-hasPermi="['inventory:material:remove']"
            >删除</el-button
          >
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页组件 -->
    <pagination
      v-show="total > 0"
      :total="total"
      :page.sync="queryParams.pageNum"
      :limit.sync="queryParams.pageSize"
      @pagination="getList"
    />

    <!-- 添加或修改物料对话框 -->
    <el-dialog :title="title" :visible.sync="open" width="900px" append-to-body>
      <el-form ref="form" :model="form" :rules="rules" label-width="100px">
        <el-row>
          <el-col :span="12">
            <el-form-item label="物料编码" prop="materialCode">
              <el-input v-model="form.materialCode" placeholder="请输入物料编码" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="物料名称" prop="materialName">
              <el-input v-model="form.materialName" placeholder="请输入物料名称" maxlength="100" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="物料类型" prop="materialType">
              <el-select
                v-model="form.materialType"
                placeholder="请选择物料类型"
                style="width: 100%"
              >
                <el-option label="原材料" value="RAW_MATERIAL" />
                <el-option label="半成品" value="SEMI_FINISHED" />
                <el-option label="成品" value="FINISHED_PRODUCT" />
                <el-option label="包装材料" value="PACKAGING" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="计量单位" prop="unit">
              <el-input v-model="form.unit" placeholder="请输入计量单位" maxlength="10" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="规格型号" prop="specification">
              <el-input v-model="form.specification" placeholder="请输入规格型号" maxlength="200" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="品牌" prop="brand">
              <el-input v-model="form.brand" placeholder="请输入品牌" maxlength="50" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item label="标准价格" prop="standardPrice">
              <el-input-number
                v-model="form.standardPrice"
                :precision="2"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="采购价格" prop="purchasePrice">
              <el-input-number
                v-model="form.purchasePrice"
                :precision="2"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="销售价格" prop="salePrice">
              <el-input-number
                v-model="form.salePrice"
                :precision="2"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item label="最小库存" prop="minStock">
              <el-input-number
                v-model="form.minStock"
                :precision="3"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="安全库存" prop="safetyStock">
              <el-input-number
                v-model="form.safetyStock"
                :precision="3"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="最大库存" prop="maxStock">
              <el-input-number
                v-model="form.maxStock"
                :precision="3"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="24">
            <el-form-item label="技术参数" prop="technicalParams">
              <el-input
                v-model="form.technicalParams"
                type="textarea"
                placeholder="请输入技术参数"
                maxlength="500"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="质量标准" prop="qualityStandard">
              <el-input
                v-model="form.qualityStandard"
                placeholder="请输入质量标准"
                maxlength="100"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitForm">确 定</el-button>
        <el-button @click="cancel">取 消</el-button>
      </div>
    </el-dialog>

    <!-- 物料详情对话框 -->
    <el-dialog title="物料详情" :visible.sync="viewOpen" width="800px" append-to-body>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="物料编码">{{ viewData.materialCode }}</el-descriptions-item>
        <el-descriptions-item label="物料名称">{{ viewData.materialName }}</el-descriptions-item>
        <el-descriptions-item label="物料类型">{{ viewData.materialType }}</el-descriptions-item>
        <el-descriptions-item label="计量单位">{{ viewData.unit }}</el-descriptions-item>
        <el-descriptions-item label="规格型号">{{ viewData.specification }}</el-descriptions-item>
        <el-descriptions-item label="品牌">{{ viewData.brand }}</el-descriptions-item>
        <el-descriptions-item label="标准价格">¥{{ viewData.standardPrice }}</el-descriptions-item>
        <el-descriptions-item label="当前库存" :span="1">
          <span :class="getStockClass(viewData.stockStatus)">
            {{ viewData.currentStock }} {{ viewData.unit }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="安全库存">{{ viewData.safetyStock }}</el-descriptions-item>
        <el-descriptions-item label="最小库存">{{ viewData.minStock }}</el-descriptions-item>
        <el-descriptions-item label="最大库存">{{ viewData.maxStock }}</el-descriptions-item>
        <el-descriptions-item label="供应商">{{ viewData.supplierName }}</el-descriptions-item>
        <el-descriptions-item label="技术参数" :span="2">{{
          viewData.technicalParams
        }}</el-descriptions-item>
        <el-descriptions-item label="质量标准" :span="2">{{
          viewData.qualityStandard
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script>
import {
  listMaterial,
  getMaterial,
  delMaterial,
  addMaterial,
  updateMaterial,
  exportMaterial
} from '@/api/inventory/material'

export default {
  name: 'Material',
  dicts: ['material_type', 'sys_normal_disable'],
  data() {
    return {
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 显示搜索条件
      showSearch: true,
      // 总条数
      total: 0,
      // 物料表格数据
      materialList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      // 是否显示详情弹出层
      viewOpen: false,
      // 详情数据
      viewData: {},
      // 查询参数
      queryParams: {
        pageNum: 1,
        pageSize: 10,
        materialCode: null,
        materialName: null,
        materialType: null,
        status: null
      },
      // 表单参数
      form: {},
      // 表单校验
      rules: {
        materialCode: [
          { required: true, message: '物料编码不能为空', trigger: 'blur' },
          { max: 50, message: '物料编码长度不能超过50个字符', trigger: 'blur' }
        ],
        materialName: [
          { required: true, message: '物料名称不能为空', trigger: 'blur' },
          { max: 100, message: '物料名称长度不能超过100个字符', trigger: 'blur' }
        ],
        materialType: [{ required: true, message: '物料类型不能为空', trigger: 'change' }],
        unit: [
          { required: true, message: '计量单位不能为空', trigger: 'blur' },
          { max: 10, message: '计量单位长度不能超过10个字符', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.getList()
  },
  methods: {
    /** 查询物料列表 */
    getList() {
      this.loading = true
      listMaterial(this.queryParams).then(response => {
        this.materialList = response.rows
        this.total = response.total
        this.loading = false
      })
    },
    // 取消按钮
    cancel() {
      this.open = false
      this.reset()
    },
    // 表单重置
    reset() {
      this.form = {
        id: null,
        materialCode: null,
        materialName: null,
        materialType: null,
        specification: null,
        unit: null,
        brand: null,
        model: null,
        standardPrice: 0,
        purchasePrice: 0,
        salePrice: 0,
        minStock: 0,
        safetyStock: 0,
        maxStock: 0,
        technicalParams: null,
        qualityStandard: null,
        version: 0
      }
      this.resetForm('form')
    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageNum = 1
      this.getList()
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm('queryForm')
      this.handleQuery()
    },
    // 多选框选中数据
    handleSelectionChange(selection) {
      this.ids = selection.map(item => item.id)
      this.single = selection.length !== 1
      this.multiple = !selection.length
    },
    /** 新增按钮操作 */
    handleAdd() {
      this.reset()
      this.open = true
      this.title = '添加物料'
    },
    /** 修改按钮操作 */
    handleUpdate(row) {
      this.reset()
      const id = row.id || this.ids
      getMaterial(id).then(response => {
        this.form = response.data
        this.open = true
        this.title = '修改物料'
      })
    },
    /** 查看详情操作 */
    handleView(row) {
      getMaterial(row.id).then(response => {
        this.viewData = response.data
        this.viewOpen = true
      })
    },
    /** 提交按钮 */
    submitForm() {
      this.$refs['form'].validate(valid => {
        if (valid) {
          if (this.form.id != null) {
            updateMaterial(this.form).then(response => {
              this.$modal.msgSuccess('修改成功')
              this.open = false
              this.getList()
            })
          } else {
            addMaterial(this.form).then(response => {
              this.$modal.msgSuccess('新增成功')
              this.open = false
              this.getList()
            })
          }
        }
      })
    },
    /** 删除按钮操作 */
    handleDelete(row) {
      const ids = row.id || this.ids
      this.$modal
        .confirm('是否确认删除物料编号为"' + ids + '"的数据项？')
        .then(function () {
          return delMaterial(ids)
        })
        .then(() => {
          this.getList()
          this.$modal.msgSuccess('删除成功')
        })
        .catch(() => {})
    },
    /** 导出按钮操作 */
    handleExport() {
      this.download(
        'inventory/material/export',
        {
          ...this.queryParams
        },
        `material_${new Date().getTime()}.xlsx`
      )
    },
    /** 库存预警 */
    handleLowStock() {
      this.$router.push('/inventory/material/lowstock')
    },
    /** 获取库存状态样式 */
    getStockClass(status) {
      switch (status) {
        case 'LOW':
          return 'stock-low'
        case 'WARNING':
          return 'stock-warning'
        case 'OVERFLOW':
          return 'stock-overflow'
        default:
          return 'stock-normal'
      }
    }
  }
}
</script>

<style scoped>
.stock-low {
  color: #f56c6c;
  font-weight: bold;
}
.stock-warning {
  color: #e6a23c;
  font-weight: bold;
}
.stock-overflow {
  color: #909399;
}
.stock-normal {
  color: #67c23a;
}
</style>
