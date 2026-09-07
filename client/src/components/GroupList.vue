<template>
  <div class="group-list">
    <div class="conv-item add" @click="$emit('create-group')">
      <span class="avatar plus">＋</span>
      <div class="conv-info"><div class="name">创建群聊</div></div>
    </div>
    <div class="conv-item" v-for="g in groups" :key="g.group_id" @click="$emit('select', g)">
      <span class="avatar" :style="avatarStyle(g.avatar_url)">{{ (g.group_name||'群').slice(0,1) }}</span>
      <div class="conv-info">
        <div class="name">{{ g.group_name }}</div>
        <div class="preview">{{ g.member_count }} 位成员</div>
      </div>
    </div>
    <div v-if="!groups.length" class="empty">暂无群聊</div>
  </div>
</template>

<script setup>
defineProps({ groups: Array })
defineEmits(['select', 'create-group'])
function avatarStyle(url) { return { background: url ? `url(${url}) center/cover` : 'var(--primary)' } }
</script>

<style scoped>
.conv-item { display:flex; align-items:center; gap:10px; padding:12px; cursor:pointer; }
.conv-item:hover { background:#f5f7fa; }
.conv-item.add:hover { background:#f0f6ff; }
.avatar { width:44px; height:44px; border-radius:50%; background:var(--primary); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:16px; }
.avatar.plus { background:#eef3ff; color:var(--primary); }
.conv-info { flex:1; min-width:0; }
.name { font-size:15px; font-weight:500; }
.preview { font-size:13px; color:var(--text-2); }
</style>