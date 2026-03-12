// ============ Theme Management ============
function getTheme() {
  const saved = localStorage.getItem('theme');
  if (saved) return saved;

  // Check system preference
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    return 'light';
  }
  return 'dark';
}

function setTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme') || getTheme();
  const next = current === 'dark' ? 'light' : 'dark';
  setTheme(next);
}

// Initialize theme on page load
(function initTheme() {
  const theme = getTheme();
  document.documentElement.setAttribute('data-theme', theme);

  // Listen for system theme changes
  if (window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (!localStorage.getItem('theme')) {
        setTheme(e.matches ? 'dark' : 'light');
      }
    });
  }
})();

// ============ Internationalization ============
const i18n = {
  zh: {
    // Common
    app_name: '应用定时管家',
    loading: '加载中...',
    save: '保存',
    cancel: '取消',
    delete: '删除',
    edit: '编辑',
    refresh: '刷新',
    back: '返回',
    logout: '登出',
    settings: '设置',

    // Header
    control_center: '控制中心',

    // Login
    login_subtitle: '登录以管理您的定时任务',
    login_with_lazycat: '使用懒猫账号登录',
    login_hint: '您需要使用懒猫微服账号登录',

    // Apps section
    installed_apps: '已安装应用',
    no_apps: '暂无已安装的应用',
    no_apps_hint: '请先在懒猫商店安装应用',
    resume_app: '恢复',
    pause_app: '休眠',
    status_running: '运行中',
    status_paused: '已休眠',
    status_stopped: '已停止',
    status_starting: '恢复中',
    status_stopping: '休眠中',
    status_error: '错误',
    multi_instance: '多实例',

    // Schedules section
    scheduled_tasks: '定时任务',
    new_task: '新建任务',
    no_schedules: '暂无定时任务',
    no_schedules_hint: '点击上方按钮创建您的第一个定时任务',
    task_name: '任务名称',
    task_name_placeholder: '例如：晚间关闭下载器',
    select_app: '选择应用',
    select_app_placeholder: '请选择应用...',
    operation: '操作',
    operation_resume: '恢复应用',
    operation_pause: '休眠应用',
    exec_time: '执行时间',
    repeat_days: '重复日期',
    days: ['日', '一', '二', '三', '四', '五', '六'],

    // Settings
    system_config: '系统配置',
    push_notification: '推送通知',
    serverchan_sendkey: 'Server酱 SendKey',
    sendkey_placeholder: '请输入 SendKey...',
    sendkey_hint: '获取 SendKey:',
    enable_notify: '启用通知',
    enable_notify_desc: '开启后将在任务执行时发送推送',
    notify_on_success: '成功时通知',
    notify_on_success_desc: '任务执行成功时发送通知',
    notify_on_failure: '失败时通知',
    notify_on_failure_desc: '任务执行失败时发送通知',
    save_settings: '保存设置',
    send_test: '发送测试',
    about: '关于',
    version: '版本',
    about_desc: '定时恢复和休眠懒猫微服上的应用，支持自定义调度计划和推送通知。',

    // Toast messages
    toast_save_success: '设置已保存',
    toast_save_failed: '保存失败',
    toast_test_sent: '测试通知已发送',
    toast_send_failed: '发送失败',
    toast_task_created: '任务创建成功',
    toast_task_updated: '任务更新成功',
    toast_task_deleted: '任务已删除',
    toast_app_resumed: '应用恢复中',
    toast_app_paused: '应用休眠中',
    toast_load_failed: '加载失败',
    toast_select_day: '请至少选择一天',
    confirm_delete: '确定要删除吗？',
    next_run: '下次执行',
    countdown_prefix: '将在',
    countdown_suffix: '后执行',
    countdown_days: '天',
    countdown_hours: '时',
    countdown_minutes: '分',
    countdown_seconds: '秒',
    countdown_disabled: '已禁用',
  },

  en: {
    app_name: 'App Scheduler',
    loading: 'Loading...',
    save: 'Save',
    cancel: 'Cancel',
    delete: 'Delete',
    edit: 'Edit',
    refresh: 'Refresh',
    back: 'Back',
    logout: 'Logout',
    settings: 'Settings',

    control_center: 'Control Center',

    login_subtitle: 'Sign in to manage your scheduled tasks',
    login_with_lazycat: 'Sign in with LazyCat',
    login_hint: 'You need to sign in with your LazyCat account',

    installed_apps: 'Installed Apps',
    no_apps: 'No apps installed',
    no_apps_hint: 'Please install apps from LazyCat Store first',
    resume_app: 'Resume',
    pause_app: 'Pause',
    status_running: 'Running',
    status_paused: 'Paused',
    status_stopped: 'Stopped',
    status_starting: 'Resuming',
    status_stopping: 'Pausing',
    status_error: 'Error',
    multi_instance: 'Multi-Instance',

    scheduled_tasks: 'Scheduled Tasks',
    new_task: 'New Task',
    no_schedules: 'No scheduled tasks',
    no_schedules_hint: 'Click the button above to create your first task',
    task_name: 'Task Name',
    task_name_placeholder: 'e.g., Stop downloader at night',
    select_app: 'Select App',
    select_app_placeholder: 'Please select an app...',
    operation: 'Operation',
    operation_resume: 'Resume App',
    operation_pause: 'Pause App',
    exec_time: 'Execution Time',
    repeat_days: 'Repeat Days',
    days: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],

    system_config: 'System Config',
    push_notification: 'Push Notification',
    serverchan_sendkey: 'ServerChan SendKey',
    sendkey_placeholder: 'Enter SendKey...',
    sendkey_hint: 'Get SendKey:',
    enable_notify: 'Enable Notifications',
    enable_notify_desc: 'Send push notifications when tasks execute',
    notify_on_success: 'Notify on Success',
    notify_on_success_desc: 'Send notification when task succeeds',
    notify_on_failure: 'Notify on Failure',
    notify_on_failure_desc: 'Send notification when task fails',
    save_settings: 'Save Settings',
    send_test: 'Send Test',
    about: 'About',
    version: 'Version',
    about_desc: 'Schedule resume and pause operations for apps on LazyCat, with custom schedules and push notifications.',

    toast_save_success: 'Settings saved',
    toast_save_failed: 'Failed to save',
    toast_test_sent: 'Test notification sent',
    toast_send_failed: 'Failed to send',
    toast_task_created: 'Task created',
    toast_task_updated: 'Task updated',
    toast_task_deleted: 'Task deleted',
    toast_app_resumed: 'Resuming app',
    toast_app_paused: 'Pausing app',
    toast_load_failed: 'Failed to load',
    toast_select_day: 'Please select at least one day',
    confirm_delete: 'Are you sure you want to delete?',
    next_run: 'Next run',
    countdown_prefix: 'in',
    countdown_suffix: '',
    countdown_days: 'd',
    countdown_hours: 'h',
    countdown_minutes: 'm',
    countdown_seconds: 's',
    countdown_disabled: 'Disabled',
  },

  ja: {
    app_name: 'アプリスケジューラー',
    loading: '読み込み中...',
    save: '保存',
    cancel: 'キャンセル',
    delete: '削除',
    edit: '編集',
    refresh: '更新',
    back: '戻る',
    logout: 'ログアウト',
    settings: '設定',

    control_center: 'コントロールセンター',

    login_subtitle: 'ログインしてタスクを管理',
    login_with_lazycat: 'LazyCatでログイン',
    login_hint: 'LazyCatアカウントでログインしてください',

    installed_apps: 'インストール済みアプリ',
    no_apps: 'アプリがインストールされていません',
    no_apps_hint: 'まずLazyCatストアからアプリをインストールしてください',
    resume_app: '再開',
    pause_app: '一時停止',
    status_running: '実行中',
    status_paused: '一時停止中',
    status_stopped: '停止済み',
    status_starting: '再開中',
    status_stopping: '一時停止処理中',
    status_error: 'エラー',
    multi_instance: 'マルチインスタンス',

    scheduled_tasks: 'スケジュールタスク',
    new_task: '新規タスク',
    no_schedules: 'タスクがありません',
    no_schedules_hint: '上のボタンをクリックして最初のタスクを作成',
    task_name: 'タスク名',
    task_name_placeholder: '例：夜間ダウンローダー停止',
    select_app: 'アプリを選択',
    select_app_placeholder: 'アプリを選択してください...',
    operation: '操作',
    operation_resume: 'アプリを再開',
    operation_pause: 'アプリを一時停止',
    exec_time: '実行時刻',
    repeat_days: '繰り返し',
    days: ['日', '月', '火', '水', '木', '金', '土'],

    system_config: 'システム設定',
    push_notification: 'プッシュ通知',
    serverchan_sendkey: 'ServerChan SendKey',
    sendkey_placeholder: 'SendKeyを入力...',
    sendkey_hint: 'SendKeyを取得:',
    enable_notify: '通知を有効化',
    enable_notify_desc: 'タスク実行時にプッシュ通知を送信',
    notify_on_success: '成功時に通知',
    notify_on_success_desc: 'タスク成功時に通知を送信',
    notify_on_failure: '失敗時に通知',
    notify_on_failure_desc: 'タスク失敗時に通知を送信',
    save_settings: '設定を保存',
    send_test: 'テスト送信',
    about: 'について',
    version: 'バージョン',
    about_desc: 'LazyCat上のアプリを定期的に再開・一時停止し、カスタムスケジュールとプッシュ通知をサポート。',

    toast_save_success: '設定を保存しました',
    toast_save_failed: '保存に失敗しました',
    toast_test_sent: 'テスト通知を送信しました',
    toast_send_failed: '送信に失敗しました',
    toast_task_created: 'タスクを作成しました',
    toast_task_updated: 'タスクを更新しました',
    toast_task_deleted: 'タスクを削除しました',
    toast_app_resumed: 'アプリを再開中',
    toast_app_paused: 'アプリを一時停止中',
    toast_load_failed: '読み込みに失敗しました',
    toast_select_day: '少なくとも1日を選択してください',
    confirm_delete: '削除してもよろしいですか？',
    next_run: '次回実行',
    countdown_prefix: '',
    countdown_suffix: '後に実行',
    countdown_days: '日',
    countdown_hours: '時間',
    countdown_minutes: '分',
    countdown_seconds: '秒',
    countdown_disabled: '無効',
  }
};

// Get current language
function getLang() {
  const saved = localStorage.getItem('lang');
  if (saved && i18n[saved]) return saved;

  const browserLang = navigator.language.split('-')[0];
  if (i18n[browserLang]) return browserLang;

  return 'zh';
}

let currentLang = getLang();

function t(key) {
  return i18n[currentLang][key] || i18n['zh'][key] || key;
}

function setLang(lang) {
  if (!i18n[lang]) return;
  currentLang = lang;
  localStorage.setItem('lang', lang);
  applyTranslations();
}

function applyTranslations() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    el.textContent = t(key);
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    el.placeholder = t(key);
  });
}

// ============ State ============
let apps = [];
let schedules = [];
let currentEditingScheduleId = null;

// ============ Init ============
document.addEventListener('DOMContentLoaded', async () => {
  applyTranslations();
  await loadUserInfo();

  // Only load data if on main page
  if (document.getElementById('appsList')) {
    await Promise.all([loadApps(), loadSchedules()]);
  }
});

// ============ User Info ============
async function loadUserInfo() {
  try {
    const resp = await fetch('/api/userinfo');
    if (!resp.ok) return;

    const user = await resp.json();
    const avatarEl = document.getElementById('userAvatar');
    const nameEl = document.getElementById('userName');
    const roleEl = document.getElementById('userRole');

    if (avatarEl) {
      if (user.avatar) {
        avatarEl.innerHTML = `<img src="${user.avatar}" alt="">`;
        avatarEl.style.background = 'none';
      } else {
        avatarEl.textContent = user.name?.charAt(0)?.toUpperCase() || 'U';
      }
    }
    if (nameEl) nameEl.textContent = user.name || user.userId;
    if (roleEl) roleEl.textContent = user.userRole || 'USER';
  } catch (err) {
    console.error('Failed to load user info:', err);
  }
}

// ============ Apps ============
async function loadApps() {
  const container = document.getElementById('appsList');
  if (!container) return;

  container.innerHTML = `<div class="loading"><div class="loading-spinner"></div></div>`;

  try {
    const resp = await fetch('/api/apps');
    if (!resp.ok) throw new Error('Failed to load apps');

    apps = await resp.json();
    renderApps();
  } catch (err) {
    console.error('Failed to load apps:', err);
    container.innerHTML = `
      <div class="empty-state" style="grid-column: 1 / -1;">
        <div class="empty-icon"><i class="ri-error-warning-line"></i></div>
        <div class="empty-title">${t('toast_load_failed')}</div>
      </div>
    `;
  }
}

function renderApps() {
  const container = document.getElementById('appsList');
  if (!container) return;

  if (apps.length === 0) {
    container.innerHTML = `
      <div class="empty-state" style="grid-column: 1 / -1;">
        <div class="empty-icon"><i class="ri-apps-line"></i></div>
        <div class="empty-title">${t('no_apps')}</div>
        <div class="empty-text">${t('no_apps_hint')}</div>
      </div>
    `;
    return;
  }

  container.innerHTML = apps.map(app => {
    const statusClass = getStatusClass(app.instanceStatus);
    const statusText = getStatusText(app.instanceStatus);
    const isRunning = app.instanceStatus === 'Status_Running';

    return `
      <div class="card app-card">
        <div class="app-icon">
          ${app.icon ? `<img src="${app.icon}" alt="${app.title}">` : '<i class="ri-app-store-line"></i>'}
        </div>
        <div class="app-info">
          <div class="app-name">
            ${app.title}${app.version ? ` <span class="app-version">v${app.version}</span>` : ''}
            ${app.multiInstance ? `<span class="app-badge app-badge-multi"><i class="ri-stack-line"></i> ${t('multi_instance')}</span>` : ''}
          </div>
          <div class="app-id">${app.appId}</div>
          <div class="app-status ${statusClass}">
            <span class="status-dot"></span>
            ${statusText}
          </div>
        </div>
        <div class="app-actions">
          ${isRunning
            ? `<button class="btn btn-danger btn-sm" onclick="pauseApp('${app.appId}')" title="${t('pause_app')}">
                <i class="ri-pause-circle-line"></i>
              </button>`
            : `<button class="btn btn-primary btn-sm" onclick="resumeApp('${app.appId}')" title="${t('resume_app')}">
                <i class="ri-play-circle-line"></i>
              </button>`
          }
        </div>
      </div>
    `;
  }).join('');

  // Update app selector in modal
  updateAppSelector();
}

function getStatusClass(status) {
  switch (status) {
    case 'Status_Running': return 'status-running';
    case 'Status_Paused':
    case 'Status_Stopped':
    case 'Status_Exited': return 'status-paused';
    case 'Status_Starting':
    case 'Status_Stopping': return 'status-paused';
    default:
      console.log('Unknown status:', status);
      return 'status-error';
  }
}

function getStatusText(status) {
  switch (status) {
    case 'Status_Running': return t('status_running');
    case 'Status_Paused': return t('status_paused');
    case 'Status_Stopped': return t('status_stopped');
    case 'Status_Exited': return t('status_stopped');
    case 'Status_Starting': return t('status_starting');
    case 'Status_Stopping': return t('status_stopping');
    default:
      console.log('Unknown status text for:', status);
      return status || t('status_error');
  }
}

async function resumeApp(appId) {
  try {
    showToast(t('toast_app_resumed'), 'info');
    const resp = await fetch(`/api/apps/${appId}/resume`, { method: 'POST' });
    if (!resp.ok) {
      const data = await resp.json();
      throw new Error(data.error);
    }
    setTimeout(loadApps, 2000);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function pauseApp(appId) {
  try {
    showToast(t('toast_app_paused'), 'info');
    const resp = await fetch(`/api/apps/${appId}/pause`, { method: 'POST' });
    if (!resp.ok) {
      const data = await resp.json();
      throw new Error(data.error);
    }
    setTimeout(loadApps, 2000);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function refreshApps() {
  loadApps();
}

function updateAppSelector() {
  const select = document.getElementById('scheduleApp');
  if (!select) return;

  const currentValue = select.value;
  select.innerHTML = `<option value="">${t('select_app_placeholder')}</option>`;

  apps.forEach(app => {
    const option = document.createElement('option');
    option.value = app.appId;
    option.textContent = app.title;
    option.dataset.title = app.title;
    select.appendChild(option);
  });

  if (currentValue) select.value = currentValue;
}

// ============ Schedules ============
async function loadSchedules() {
  const container = document.getElementById('schedulesList');
  if (!container) return;

  container.innerHTML = `<div class="loading"><div class="loading-spinner"></div></div>`;

  try {
    const resp = await fetch('/api/schedules');
    if (!resp.ok) throw new Error('Failed to load schedules');

    schedules = await resp.json();
    renderSchedules();
  } catch (err) {
    console.error('Failed to load schedules:', err);
    container.innerHTML = `
      <div class="empty-state" style="grid-column: 1 / -1;">
        <div class="empty-icon"><i class="ri-error-warning-line"></i></div>
        <div class="empty-title">${t('toast_load_failed')}</div>
      </div>
    `;
  }
}

// Calculate next execution time for a schedule
function getNextExecutionTime(schedule) {
  if (!schedule.enabled || !schedule.weekDays || schedule.weekDays.length === 0) {
    return null;
  }

  const now = new Date();
  const currentDay = now.getDay();
  const currentHour = now.getHours();
  const currentMinute = now.getMinutes();

  // Check each day in the next 7 days
  for (let dayOffset = 0; dayOffset < 7; dayOffset++) {
    const checkDay = (currentDay + dayOffset) % 7;

    if (schedule.weekDays.includes(checkDay)) {
      const nextExecution = new Date(now);
      nextExecution.setDate(now.getDate() + dayOffset);
      nextExecution.setHours(schedule.hour, schedule.minute, 0, 0);

      // If it's today, check if the time hasn't passed yet
      if (dayOffset === 0) {
        if (currentHour < schedule.hour || (currentHour === schedule.hour && currentMinute < schedule.minute)) {
          return nextExecution;
        }
      } else {
        return nextExecution;
      }
    }
  }

  return null;
}

// Format countdown display
function formatCountdown(nextExecution) {
  if (!nextExecution) {
    return `<span class="countdown-disabled">${t('countdown_disabled')}</span>`;
  }

  const now = new Date();
  const diff = nextExecution - now;

  if (diff <= 0) {
    return `<span class="countdown-now">即将执行...</span>`;
  }

  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);

  const parts = [];
  if (days > 0) parts.push(`${days}${t('countdown_days')}`);
  if (hours > 0 || days > 0) parts.push(`${hours}${t('countdown_hours')}`);
  if (minutes > 0 || hours > 0 || days > 0) parts.push(`${minutes}${t('countdown_minutes')}`);
  if (days === 0 && hours === 0) parts.push(`${seconds}${t('countdown_seconds')}`);

  const timeText = parts.join(' ');
  const prefix = t('countdown_prefix');
  const suffix = t('countdown_suffix');

  // Build the complete message with proper spacing
  let message = '';
  if (prefix) message += prefix + ' ';
  message += timeText;
  if (suffix) message += ' ' + suffix;

  return message;
}

// Update all countdowns
function updateCountdowns() {
  schedules.forEach(sch => {
    const element = document.getElementById(`countdown-${sch.id}`);
    if (element) {
      const nextExecution = getNextExecutionTime(sch);
      element.innerHTML = formatCountdown(nextExecution);
    }
  });
}

// Start countdown timer
let countdownInterval = null;
function startCountdownTimer() {
  if (countdownInterval) {
    clearInterval(countdownInterval);
  }
  countdownInterval = setInterval(updateCountdowns, 1000);
}

function renderSchedules() {
  const container = document.getElementById('schedulesList');
  if (!container) return;

  if (schedules.length === 0) {
    container.innerHTML = `
      <div class="empty-state" style="grid-column: 1 / -1;">
        <div class="empty-icon"><i class="ri-calendar-schedule-line"></i></div>
        <div class="empty-title">${t('no_schedules')}</div>
        <div class="empty-text">${t('no_schedules_hint')}</div>
      </div>
    `;
    if (countdownInterval) {
      clearInterval(countdownInterval);
      countdownInterval = null;
    }
    return;
  }

  const dayNames = t('days');

  container.innerHTML = schedules.map(sch => {
    const nextExecution = getNextExecutionTime(sch);
    const countdownText = formatCountdown(nextExecution);

    return `
    <div class="card schedule-card">
      <div class="schedule-header">
        <div>
          <div class="schedule-name">${sch.name}</div>
          <div class="schedule-app">${sch.appTitle || sch.appId}</div>
        </div>
        <label class="toggle schedule-toggle">
          <input type="checkbox" ${sch.enabled ? 'checked' : ''} onchange="toggleSchedule('${sch.id}')">
          <span class="toggle-slider"></span>
        </label>
      </div>
      <div class="schedule-countdown">
        <i class="ri-timer-line"></i>
        <span id="countdown-${sch.id}">${countdownText}</span>
      </div>
      <div class="schedule-body">
        <div class="schedule-time">
          <i class="ri-time-line schedule-time-icon"></i>
          ${String(sch.hour).padStart(2, '0')}:${String(sch.minute).padStart(2, '0')}
        </div>
        <div class="schedule-operation ${sch.operation}">
          <i class="ri-${sch.operation === 'resume' ? 'play' : 'pause'}-circle-line"></i>
          ${sch.operation === 'resume' ? t('operation_resume') : t('operation_pause')}
        </div>
      </div>
      <div class="schedule-days">
        ${[0, 1, 2, 3, 4, 5, 6].map(day => `
          <span class="schedule-day ${sch.weekDays.includes(day) ? 'active' : ''}">${dayNames[day]}</span>
        `).join('')}
      </div>
      <div class="schedule-footer">
        <button class="btn btn-secondary btn-sm" onclick="editSchedule('${sch.id}')">
          <i class="ri-edit-line"></i>
          ${t('edit')}
        </button>
        <button class="btn btn-danger btn-sm" onclick="deleteSchedule('${sch.id}')">
          <i class="ri-delete-bin-line"></i>
          ${t('delete')}
        </button>
      </div>
    </div>
  `;
  }).join('');

  // Start countdown timer
  startCountdownTimer();
}

function openScheduleModal(scheduleId = null) {
  currentEditingScheduleId = scheduleId;
  const modal = document.getElementById('scheduleModal');
  const title = document.getElementById('scheduleModalTitle');
  const form = document.getElementById('scheduleForm');

  title.textContent = scheduleId ? t('edit') + ' ' + t('scheduled_tasks') : t('new_task');
  form.reset();

  // Reset day checkboxes
  document.querySelectorAll('input[name="weekDays"]').forEach(cb => cb.checked = false);

  if (scheduleId) {
    const sch = schedules.find(s => s.id === scheduleId);
    if (sch) {
      document.getElementById('scheduleName').value = sch.name;
      document.getElementById('scheduleApp').value = sch.appId;
      document.getElementById('scheduleOperation').value = sch.operation;
      document.getElementById('scheduleHour').value = sch.hour;
      document.getElementById('scheduleMinute').value = sch.minute;
      sch.weekDays.forEach(day => {
        const cb = document.querySelector(`input[name="weekDays"][value="${day}"]`);
        if (cb) cb.checked = true;
      });
    }
  }

  modal.classList.add('active');
}

function closeScheduleModal() {
  document.getElementById('scheduleModal').classList.remove('active');
  currentEditingScheduleId = null;
}

function editSchedule(id) {
  openScheduleModal(id);
}

async function saveSchedule(event) {
  event.preventDefault();

  const weekDays = Array.from(document.querySelectorAll('input[name="weekDays"]:checked'))
    .map(cb => parseInt(cb.value));

  if (weekDays.length === 0) {
    showToast(t('toast_select_day'), 'warning');
    return;
  }

  const appSelect = document.getElementById('scheduleApp');
  const selectedOption = appSelect.options[appSelect.selectedIndex];

  const data = {
    name: document.getElementById('scheduleName').value,
    appId: appSelect.value,
    appTitle: selectedOption?.dataset?.title || appSelect.value,
    operation: document.getElementById('scheduleOperation').value,
    hour: parseInt(document.getElementById('scheduleHour').value),
    minute: parseInt(document.getElementById('scheduleMinute').value),
    weekDays: weekDays
  };

  try {
    let resp;
    if (currentEditingScheduleId) {
      resp = await fetch(`/api/schedules/${currentEditingScheduleId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
    } else {
      resp = await fetch('/api/schedules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
    }

    if (!resp.ok) {
      const err = await resp.json();
      throw new Error(err.error);
    }

    showToast(currentEditingScheduleId ? t('toast_task_updated') : t('toast_task_created'), 'success');
    closeScheduleModal();
    loadSchedules();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function deleteSchedule(id) {
  if (!confirm(t('confirm_delete'))) return;

  try {
    const resp = await fetch(`/api/schedules/${id}`, { method: 'DELETE' });
    if (!resp.ok) {
      const data = await resp.json();
      throw new Error(data.error);
    }
    showToast(t('toast_task_deleted'), 'success');
    loadSchedules();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function toggleSchedule(id) {
  try {
    const resp = await fetch(`/api/schedules/${id}/toggle`, { method: 'POST' });
    if (!resp.ok) {
      const data = await resp.json();
      throw new Error(data.error);
    }
    loadSchedules();
  } catch (err) {
    showToast(err.message, 'error');
    loadSchedules();
  }
}

// ============ Toast ============
function showToast(message, type = 'info') {
  const container = document.getElementById('toastContainer');
  if (!container) return;

  const icons = {
    success: 'ri-check-line',
    error: 'ri-error-warning-line',
    warning: 'ri-alert-line',
    info: 'ri-information-line'
  };

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <i class="toast-icon ${icons[type]}"></i>
    <span class="toast-message">${message}</span>
  `;

  container.appendChild(toast);

  setTimeout(() => {
    toast.classList.add('removing');
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

// ============ Keyboard shortcuts ============
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    closeScheduleModal();
  }
});

// Close modal on overlay click
document.addEventListener('click', (e) => {
  if (e.target.classList.contains('modal-overlay')) {
    closeScheduleModal();
  }
});
