<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="group-modal-title">
      <div class="m-title" id="group-modal-title">创建群聊</div>
      <label for="group-name" class="sr-only">群名称</label>
      <input id="group-name" class="input" v-model="name" placeholder="群名称(必填,≤30字符)" />
      <label for="group-ann" class="sr-only">群公告</label>
      <input id="group-ann" class="input" v-model="announcement" placeholder="群公告(选填)" style="margin-top:8px" />
      <div class="pick-title">选择好友(至少1位)</div>
      <div class="pick-list">
        <button v-for="f in friends" :key="f.user_id" class="pick-item"
          :class="{checked: selected.includes(f.user_id)}" @click="toggle(f.user_id)" type="button">
          <span class="avatar" aria-hidden="true">{{ (f.nickname||'?').slice(0,1) }}</span>
          <span class="f-name">{{ f.remark || f.nickname }}</span>
          <span class="check" aria-hidden="true">{{ selected.includes(f.user_id) ? '✓' : '' }}</span>
        </button>
        <p v-if="!friends.length" class="empty">暂无好友，请先添加好友</p>
      </div>
      <div class="m-ops">
        <button class="btn" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="submit" type="button">创建</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api/http'

const props = defineProps({ friends: Array })
const emit = defineEmits(['close', 'created'])

const name = ref('')
const announcement = ref('')
const selected = ref([])

function toggle(id) {
  const i = selected.value.indexOf(id)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(id)
}
async function submit() {
  if (!name.value) return alert('请填写群名称')
  try {
    const { data } = await api.post('/groups', {
      name: name.value,
      announcement: announcement.value,
      member_ids: selected.value
    })
    emit('created', data)
  } catch (err) { alert(err.message) }
}
</script>

<style scoped>
.sr-only { position:absolute; width:1px; height:1px; margin:-1px; padding:0; clip:rect(0,0,0,0); border:0; overflow:hidden; white-space:nowrap; }
.modal-mask { position:fixed; inset:0; background:rgba(0,0,0,.4); display:flex;
  align-items:center; justify-content:center; z-index:200; }
.modal { width:420px; background:#fff; border-radius:12px; padding:20px; }
.m-title { font-size:17px; font-weight:600; margin-bottom:14px; }
.pick-title { font-size:13px; color:var(--text-2); margin:12px 0 6px; }
.pick-list { max-height:260px; overflow:auto; border:1px solid var(--border); border-radius:8px; padding:6px; }
.pick-item { display:flex; align-items:center; gap:10px; padding:8px; cursor:pointer; border-radius:6px;
  width:100%; text-align:left; border:none; background:transparent; font:inherit; color:inherit; }
.pick-item:hover { background:#f5f7fa; }
.pick-item:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; }
.pick-item.checked { background:#eef6ff; }
.avatar { width:34px; height:34px; border-radius:50%; background:var(--primary); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:14px; flex-shrink:0; }
.f-name { flex:1; font-size:14px; }
.check { width:20px; text-align:center; color:var(--primary); font-weight:700; }
.m-ops { display:flex; justify-content:flex-end; gap:10px; margin-top:16px; }
</style>