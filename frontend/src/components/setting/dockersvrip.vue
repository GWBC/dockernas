<template>
  <div>
    <div class="card_style">
      <div class="center_div" style="padding-top: 10px; padding-bottom: 10px">
        <div style="flex-grow: 0.98"></div>
        <el-button type="primary" class="big_button" @click="createSvrInfo">添加服务</el-button>
      </div>

      <el-table :data="svrinfos" @row-dblclick="editInfo" stripe class="table_css" :row-style="{ height: '50px' }"
        :cell-style="{ padding: '0px' }" :row-key="row => { return row.id }">
        <el-table-column prop="name" label="名称" sortable min-width="30%" />
        <el-table-column prop="ip" label="容器服务地址" sortable min-width="30%" />
        <el-table-column prop="use" label="是否使用" sortable min-width="30%">
          <template #default="scope">
            <span v-show="scope.row.use==0" style="color:rgb(224, 40, 40)">否</span>
            <span v-show="scope.row.use==1" style="color: rgb(53, 163, 53);">是</span>
            <!-- <span style="color:red"> {{ toStrUse(scope.row.use) }} </span> -->
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180px" #default="scope">
          <el-button size="small" type="success" @click="editInfo(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="delInfo(scope.row)">删除</el-button>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogInfo.visible" :title="dialogInfo.title" class="small_dialog">
      <div class="center_div">
        <div>
          <div class="input_div">
            <div class="first_input">服务名</div>
            <div>
              <el-input class="big_input" size="large" v-model="dialogInfo.name">
              </el-input>
            </div>
          </div>
          <div class="input_div">
            <div class="first_input">服务地址</div>
            <div>
              <el-input class="big_input" size="large" v-model="dialogInfo.ip">
              </el-input>
            </div>
          </div>
          <div class="input_div">
            <div class="first_input">是否激活</div>
            <div>
              <el-select class="big_input" size="large" v-model="dialogInfo.use" placeholder="请选择">
                <el-option v-for="item in dialogInfo.options" :key="item.value" :label="item.label" :value="item.value">
                </el-option>
              </el-select>
            </div>
          </div>
          <div class="center_div" style="margin-top: 50px">
            <el-button type="primary" style="height: 40px; width: 200px" @click="replaceSvrInfo">{{ $t("common.yes")
            }}</el-button>
          </div>
        </div>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import { getDockerSvrInfos, addDockerSvrInfo, editDockerSvrInfo, delDockerSvrInfo } from "../../api/dockersvrip";

export default {
  name: "dockersvrip",
  methods: {
    flush() {
      getDockerSvrInfos().then((response) => {
        this.svrinfos = response.data.list
      });
    },
    toStrUse(index) {
      if (index == 0) {
        return "否"
      }

      if (index == 1) {
        return "是"
      }
    },
    delInfo(row) {
      delDockerSvrInfo(row.id).then((response) => {
        this.flush()
      })
    },
    editInfo(row) {
      this.dialogInfo.title = "编辑服务信息"
      this.dialogInfo.id = row.id
      this.dialogInfo.name = row.name
      this.dialogInfo.ip = row.ip
      this.dialogInfo.use = row.use
      this.dialogInfo.visible = true
    },
    createSvrInfo() {
      this.dialogInfo.title = "添加服务信息"
      this.dialogInfo.id = undefined
      this.dialogInfo.name = ""
      this.dialogInfo.ip = ""
      this.dialogInfo.use = 1
      this.dialogInfo.visible = true
    },
    replaceSvrInfo() {
      if (this.dialogInfo.id !== undefined) {
        if (this.checkDataValid(this.dialogInfo, this.dialogInfo.id)) {
          editDockerSvrInfo(this.dialogInfo).then((response) => {
            this.flush()
            this.dialogInfo.visible = false
          })
        }
      } else {
        if (this.checkDataValid(this.dialogInfo)) {
          addDockerSvrInfo(this.dialogInfo).then((response) => {
            this.flush()
            this.dialogInfo.visible = false
          })
        }
      }
    },
    checkDataValid(info, filterRowId) {
      for (let index in this.svrinfos) {
        if (filterRowId != undefined && this.svrinfos[index].id == filterRowId) {
          continue
        }

        if (info.name == this.svrinfos[index].name) {
          this.$message.error('name <' + info.name + '> exist');
          return false
        } else if (info.ip == this.svrinfos[index].ip) {
          this.$message.error('address <' + info.ip + '> exist');
          return false
        }
      }

      return true
    }
  },
  data() {
    return {
      dialogInfo: {
        id: 0,
        title: "",
        name: "",
        ip: "",
        use: 0,
        visible: false,
        options: [{ value: 1, label: "是" }, { value: 0, label: "否" }],
      },
      svrinfos: []
    }
  },
  mounted() {
    this.flush();
  },
};
</script>

<style>
@import "../../css/common.css";
@import "../../css/picture.css";
@import "../../css/menu.css";
@import "../../css/text.css";
@import "../../css/dialog.css";
</style>