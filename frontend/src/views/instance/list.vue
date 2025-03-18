<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getInstanceList,
  saveInstance,
  type InfoEntity,
  type Instance
} from "@/api/instance";
import SignatureTest from "./components/SignatureTest.vue";

defineOptions({
  name: "InstanceList"
});

const loading = ref(false);
const total = ref(0);
const page = ref(1);
const size = ref(50);
const dataList = ref<Instance[]>([]);

const dialogVisible = ref(false);
const dialogTitle = ref("");
const formData = ref<InfoEntity>({
  id: 0,
  inst_type: 0,
  inner_ip_v_6: "",
  telecom_ip_v_6: "",
  unicom_ip_v_6: "",
  mobile_ip_v_6: "",
  inner_ip_v_4: "",
  telecom_ip_v_4: "",
  unicom_ip_v_4: "",
  mobile_ip_v_4: "",
  status: "",
  desc: ""
});

const statusOptions = [
  { label: "允许分配", value: "allow_alloc" },
  { label: "允许调试", value: "allow_debug" },
  { label: "已下架", value: "offline" }
];

const typeOptions = [
  { label: "房间实例", value: 0 },
  { label: "衣服实例", value: 1 },
  { label: "人物实例", value: 2 }
];

const getList = async () => {
  try {
    loading.value = true;
    const res = await getInstanceList({
      page: page.value,
      size: size.value
    });
    if (res.code === 0) {
      dataList.value = res.data.instances;
      total.value = res.data.total;
    } else {
      ElMessage.error(res.message || "获取实例列表失败");
    }
  } catch (error) {
    console.error("获取实例列表失败:", error);
    ElMessage.error("获取实例列表失败");
  } finally {
    loading.value = false;
  }
};

const handleSizeChange = (val: number) => {
  size.value = val;
  getList();
};

const handleCurrentChange = (val: number) => {
  page.value = val;
  getList();
};

const handleAdd = () => {
  dialogTitle.value = "新增实例";
  formData.value = {
    id: 0,
    inst_type: 0,
    inner_ip_v_6: "",
    telecom_ip_v_6: "",
    unicom_ip_v_6: "",
    mobile_ip_v_6: "",
    inner_ip_v_4: "",
    telecom_ip_v_4: "",
    unicom_ip_v_4: "",
    mobile_ip_v_4: "",
    status: "",
    desc: ""
  };
  dialogVisible.value = true;
};

const handleEdit = (row: Instance) => {
  dialogTitle.value = "编辑实例";
  formData.value = { ...row.info_entity };
  dialogVisible.value = true;
};

const handleDelete = (row: Instance) => {
  ElMessageBox.confirm("确认删除该实例吗？", "提示", {
    type: "warning"
  })
    .then(() => {
      ElMessage.success("删除成功");
      getList();
    })
    .catch(() => {});
};

const handleSubmit = async () => {
  try {
    const res = await saveInstance({ instance: formData.value });
    if (res.code === 0) {
      ElMessage.success("保存成功");
      dialogVisible.value = false;
      getList();
    } else {
      ElMessage.error(res.message || "保存失败");
    }
  } catch (error) {
    console.error("保存实例失败:", error);
    ElMessage.error("保存实例失败");
  }
};

onMounted(() => {
  getList();
});
</script>

<template>
  <div class="main">
    <!-- <SignatureTest /> -->

    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="font-medium">实例列表</span>
          <el-button type="primary" @click="handleAdd">新增实例</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="dataList" border style="width: 100%">
        <el-table-column prop="info_entity.id" label="ID" width="80">
          <template #default="{ row }">
            {{ row.info_entity.id }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.inst_type"
          label="实例类型"
          width="100"
        >
          <template #default="{ row }">
            {{
              typeOptions.find(t => t.value === row.info_entity.inst_type)
                ?.label
            }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.inner_ip_v_4"
          label="内网IPv4"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.info_entity.inner_ip_v_4 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.telecom_ip_v_4"
          label="电信IPv4"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.info_entity.telecom_ip_v_4 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.unicom_ip_v_4"
          label="联通IPv4"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.info_entity.unicom_ip_v_4 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.mobile_ip_v_4"
          label="移动IPv4"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.info_entity.mobile_ip_v_4 }}
          </template>
        </el-table-column>
        <!--
        <el-table-column
          prop="info_entity.inner_ip_v_6"
          label="内网IPv6"
          min-width="200"
        >
          <template #default="{ row }">
            {{ row.info_entity.inner_ip_v_6 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.telecom_ip_v_6"
          label="电信IPv6"
          min-width="200"
        >
          <template #default="{ row }">
            {{ row.info_entity.telecom_ip_v_6 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.unicom_ip_v_6"
          label="联通IPv6"
          min-width="200"
        >
          <template #default="{ row }">
            {{ row.info_entity.unicom_ip_v_6 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_entity.mobile_ip_v_6"
          label="移动IPv6"
          min-width="200"
        >
          <template #default="{ row }">
            {{ row.info_entity.mobile_ip_v_6 }}
          </template>
        </el-table-column>
      -->
        <el-table-column prop="info_entity.status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="
                row.info_entity.status === 'offline'
                  ? 'danger'
                  : row.info_entity.status === 'allow_debug'
                    ? 'warning'
                    : 'success'
              "
            >
              {{
                statusOptions.find(s => s.value === row.info_entity.status)
                  ?.label
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分配状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.info_real_time.alloced ? 'warning' : 'success'">
              {{ row.info_real_time.alloced ? "已分配" : "未分配" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="实例状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.info_real_time.status_code ? 'danger' : 'success'"
            >
              {{ row.info_real_time.status_code ? "异常" : "正常" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="info_real_time.room_id"
          label="房间ID"
          min-width="180"
        >
          <template #default="{ row }">
            {{
              row.info_real_time.room_id
                ? String(row.info_real_time.room_id)
                : "-"
            }}
          </template>
        </el-table-column>
        <el-table-column
          prop="info_real_time.status"
          label="上报错误信息"
          min-width="150"
        >
          <template #default="{ row }">
            {{ row.info_real_time.status }}
          </template>
        </el-table-column>
        <el-table-column prop="info_entity.desc" label="描述" min-width="150">
          <template #default="{ row }">
            {{ row.info_entity.desc }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)"
              >编辑</el-button
            >
            <el-button link type="danger" @click="handleDelete(row)"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="800px"
      destroy-on-close
    >
      <el-form :model="formData" label-width="120px">
        <el-form-item label="实例类型">
          <el-select v-model="formData.inst_type">
            <el-option
              v-for="item in typeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="内网IPv4">
          <el-input v-model="formData.inner_ip_v_4" />
        </el-form-item>
        <el-form-item label="电信IPv4">
          <el-input v-model="formData.telecom_ip_v_4" />
        </el-form-item>
        <el-form-item label="联通IPv4">
          <el-input v-model="formData.unicom_ip_v_4" />
        </el-form-item>
        <el-form-item label="移动IPv4">
          <el-input v-model="formData.mobile_ip_v_4" />
        </el-form-item>
        <el-form-item label="内网IPv6">
          <el-input v-model="formData.inner_ip_v_6" />
        </el-form-item>
        <el-form-item label="电信IPv6">
          <el-input v-model="formData.telecom_ip_v_6" />
        </el-form-item>
        <el-form-item label="联通IPv6">
          <el-input v-model="formData.unicom_ip_v_6" />
        </el-form-item>
        <el-form-item label="移动IPv6">
          <el-input v-model="formData.mobile_ip_v_6" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="formData.status">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.desc" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.main {
  padding: 20px;

  .box-card {
    margin-bottom: 20px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
