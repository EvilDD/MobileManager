<script setup lang="ts">
import { Platform, Link, VideoPlay, VideoCamera, User, Warning } from '@element-plus/icons-vue';

defineOptions({
  name: "Welcome"
});
</script>

<template>
  <div class="welcome-container">
    <div class="welcome-content">
      <h1 class="welcome-title">欢迎使用 <span class="highlight">Mobile Manager</span></h1>
      
      <div class="card disclaimer-card">
        <h2 class="card-title">
          <el-icon><Warning /></el-icon>
          免责声明
        </h2>
        <p class="disclaimer-text">
          本工具仅供合法用途使用，不得用于任何非法活动。常用于以下合法场景：
        </p>
        <ul class="disclaimer-list">
          <li>自动化测试</li>
          <li>并发测试</li>
          <li>搭建云手机平台</li>
          <li>云游戏测试</li>
          <li>APP兼容性测试</li>
          <li>学习研究用途</li>
        </ul>
        <p class="disclaimer-text">
          请遵守相关法律法规，合理使用本工具。用户需对使用过程中的行为负全部责任。
        </p>
      </div>
      
      <div class="card">
        <h2 class="card-title">
          <el-icon><Platform /></el-icon>
          项目地址
        </h2>
        <a href="https://github.com/EvilDD/MobileManager" target="_blank" class="link">
          <el-icon><Link /></el-icon>
          GitHub - MobileManager
        </a>
      </div>
      
      <div class="card">
        <h2 class="card-title">
          <el-icon><VideoPlay /></el-icon>
          相关视频
        </h2>
        <a href="https://www.bilibili.com/video/BV1bL5jz7E6y" target="_blank" class="link">
          <el-icon><VideoCamera /></el-icon>
          群控云手机模拟器自动化-合集_哔哩哔哩
        </a>
      </div>
      
      <div class="card">
        <h2 class="card-title">
          <el-icon><User /></el-icon>
          关于博主
        </h2>
        <p class="about-text">欢迎关注博主获取更多优质内容！</p>
        <a href="https://space.bilibili.com/25110244?spm_id_from=333.1007.0.0" target="_blank">
          <el-button type="primary" class="follow-button">关注博主</el-button>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.welcome-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 120px);
  padding: 2rem;
  background-color: var(--el-bg-color);
}

.welcome-content {
  max-width: 800px;
  width: 100%;
}

.welcome-title {
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 2.5rem;
  color: var(--el-text-color-primary);
}

.highlight {
  color: var(--el-color-primary);
  font-weight: 700;
}

.card {
  background: var(--el-bg-color-overlay);
  border-radius: 8px;
  box-shadow: var(--el-box-shadow-light);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  transition: transform 0.3s, box-shadow 0.3s;
}

.card:hover {
  transform: translateY(-5px);
  box-shadow: var(--el-box-shadow);
}

.disclaimer-card {
  border-left: 4px solid var(--el-color-warning);
}

.disclaimer-text {
  color: var(--el-text-color-regular);
  font-size: 1rem;
  margin-bottom: 0.5rem;
}

.disclaimer-list {
  color: var(--el-text-color-regular);
  padding-left: 1.5rem;
  margin-bottom: 0.5rem;
}

.disclaimer-list li {
  margin-bottom: 0.25rem;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--el-color-primary);
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 1.5rem;
}

.link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--el-color-primary);
  text-decoration: none;
  font-size: 1.1rem;
  padding: 0.5rem;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.link:hover {
  background-color: var(--el-color-primary-light-9);
}

.about-text {
  color: var(--el-text-color-regular);
  font-size: 1.1rem;
  margin-bottom: 1rem;
}

.follow-button {
  width: 100%;
}

@media (max-width: 768px) {
  .welcome-title {
    font-size: 1.8rem;
  }
  
  .card {
    padding: 1rem;
  }
}
</style>
