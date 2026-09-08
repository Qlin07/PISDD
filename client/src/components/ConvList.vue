<template>
  <div class="conv-list">
    <button class="conv-item add" @click="$emit('create-group')" type="button">
      <span class="avatar plus">＋</span>
      <div class="conv-info"><div class="name">发起群聊</div></div>
    </button>
    <button class="conv-item" v-for="c in conversations" :key="c.conversation_id"
      :class="{active: current && current.conversation_id===c.conversation_id}"
      @click="$emit('select', c)" type="button">
      <span class="avatar" :style="avatarStyle(c.avatar)">{{ (c.display_name||'?').slice(0,1) }}</span>
      <div class="conv-info">
        <div class="name-row">
          <span class="name">{{ c.display_name }}</span>
          <span class="time">{{ fmtTime(c.last_msg_time) }}</span>
        </div>
        <div class="preview-row">
          <span class="preview">{{ c.last_msg_preview || (c.type===0?'[单聊]':'[群聊]') }}</span>
          <span class="unread" v-if="c.unread_count > 0" aria-live="polite">{{ c.unread_count > 99 ? '99+' : c.unread_count }}</span>
          <span class="tag" v-if="c.is_top">置顶</span>
          <span class="tag mute" v-if="c.is_mute">免扰</span>
        </div>
      </div>
    </button>
    <p v-if="!conversations.length" class="empty">还没有消息，点击「发起群聊」或去联系人找人聊聊</p>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
defineProps({ conversations: Array, current: Object })
defineEmits(['select', 'create-group'])

function avatarStyle(url) {
  return { background: url ? `url(${url}) center/cover` : 'var(--primary)' }
}
function fmtTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
  }
  return `${d.getMonth()+1}/${d.getDate()}`
}
</script>

<style scoped>
.conv-item { display:flex; align-items:center; gap:10px; padding:12px; cursor:pointer;
  width:100%; text-align:left; border:none; background:transparent; font:inherit; color:inherit;
  transition: background .15s; border-radius:8px; }
.conv-item:hover { background:var(--surface-3); }
.conv-item:focus-visible { outline:2px solid var(--primary); outline-offset:-2px; background:var(--surface-3); }
.conv-item.active { background:var(--surface-2); box-shadow: inset 3px 0 0 var(--primary); }
.conv-item.add:hover { background:var(--surface-3); }
.avatar { width:44px; height:44px; border-radius:50%; background:linear-gradient(135deg,var(--primary),var(--primary-dim));
  color:#fff; display:flex; align-items:center; justify-content:center; font-size:15px; flex-shrink:0; font-weight:600; box-shadow: 0 2px 8px var(--primary-glow); }
.avatar.plus { background:var(--surface-3); color:var(--primary); box-shadow:none; }
.conv-info { flex:1; min-width:0; }
.name-row { display:flex; justify-content:space-between; align-items:center; }
.name { font-size:15px; font-weight:500; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--text); }
.time { font-size:12px; color:var(--text-3); }
.preview-row { display:flex; align-items:center; gap:6px; margin-top:2px; }
.preview { font-size:13px; color:var(--text-2); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; flex:1; }
.unread { background:var(--danger); color:#fff; font-size:11px; border-radius:10px; padding:1px 6px; min-width:20px; text-align:center; box-shadow:0 0 8px var(--danger-glow); }
.tag { font-size:10px; color:var(--primary); border:1px solid var(--primary-dim); border-radius:4px; padding:0 3px; }
.tag.mute { color:var(--text-3); border-color:var(--border); }
</style>