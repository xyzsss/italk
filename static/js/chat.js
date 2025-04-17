// 获取DOM元素
const messagesContainer = document.getElementById('messages');
const messageInput = document.getElementById('message-input');
const sendButton = document.getElementById('send-btn');
const userIP = document.getElementById('user-ip');
const usernameInput = document.getElementById('username-input');
const setUsernameButton = document.getElementById('set-username-btn');
const imageUpload = document.getElementById('image-upload');
const fileUpload = document.getElementById('file-upload');
const emojiToggle = document.getElementById('emoji-toggle');
const emojiPanel = document.getElementById('emoji-panel');
const emojiList = document.querySelector('.emoji-list');
const userList = document.getElementById('user-list');
const statsContent = document.getElementById('stats-content');
const searchInput = document.getElementById('search-input');
const searchButton = document.getElementById('search-btn');
const searchResults = document.getElementById('search-results');
const tabButtons = document.querySelectorAll('.tab-btn');
const tabContents = document.querySelectorAll('.tab-content');

// WebSocket连接
let socket;
let currentUserIP = '';
let localUserID = null;
let messageMap = new Map(); // 存储消息ID和DOM元素的映射

// 消息类型
const MESSAGE_TYPES = {
    TEXT: 'text',
    IMAGE: 'image',
    EMOJI: 'emoji',
    SYSTEM: 'system',
    USER: 'user',
    USERS: 'users',
    STATS: 'stats',
    FILE: 'file',
    RECALL: 'recall'
};

// 文件大小格式化
function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
}

// 初始化应用
function init() {
    // 获取用户IP
    fetchUserIP();
    
    // 初始化WebSocket连接
    initWebSocket();
    
    // 初始化表情列表
    initEmojis();
    
    // 初始化历史消息
    fetchMessages();
    
    // 绑定事件
    bindEvents();
    
    // 每60秒刷新一次在线用户列表
    setInterval(fetchOnlineUsers, 60000);
}

// 获取用户IP
function fetchUserIP() {
    // 通过服务端获取客户端真实IP
    fetch('/api/users/online')
        .then(response => response.json())
        .then(users => {
            if (users && users.length > 0) {
                // 假设第一个返回的是当前用户
                currentUserIP = users[0].ip;
                userIP.textContent = `IP: ${currentUserIP}`;
            }
        })
        .catch(error => console.error('获取IP失败:', error));
}

// 初始化WebSocket连接
function initWebSocket() {
    // 构建WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    // 创建WebSocket连接
    socket = new WebSocket(wsUrl);
    
    // WebSocket事件
    socket.onopen = () => {
        console.log('WebSocket 连接已建立');
        
        // 连接成功后获取在线用户和统计信息
        fetchOnlineUsers();
        fetchStats();
    };
    
    socket.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            handleMessage(message);
        } catch (error) {
            console.error('解析消息失败:', error);
        }
    };
    
    socket.onclose = () => {
        console.log('WebSocket 连接已关闭');
        // 尝试重新连接
        setTimeout(initWebSocket, 3000);
    };
    
    socket.onerror = (error) => {
        console.error('WebSocket 错误:', error);
    };
}

// 处理接收到的消息
function handleMessage(message) {
    switch (message.type) {
        case MESSAGE_TYPES.TEXT:
        case MESSAGE_TYPES.IMAGE:
        case MESSAGE_TYPES.EMOJI:
        case MESSAGE_TYPES.FILE:
            renderMessage(message);
            break;
        case MESSAGE_TYPES.SYSTEM:
            renderSystemMessage(message);
            break;
        case MESSAGE_TYPES.USER:
            // 用户信息更新
            if (message.user_id === localUserID) {
                usernameInput.value = message.username;
            }
            break;
        case MESSAGE_TYPES.USERS:
            // 更新在线用户列表
            if (message.data) {
                renderUserList(message.data);
            }
            break;
        case MESSAGE_TYPES.STATS:
            // 更新统计信息
            if (message.data) {
                renderStats(message.data);
            }
            break;
        case MESSAGE_TYPES.RECALL:
            // 处理消息撤回
            handleRecalledMessage(message);
            break;
    }
    
    // 滚动到底部
    scrollToBottom();
}

// 渲染普通消息
function renderMessage(message) {
    const isOwnMessage = message.ip === currentUserIP;
    const messageElement = document.createElement('div');
    messageElement.className = `message ${isOwnMessage ? 'own-message' : 'user-message'}`;
    messageElement.dataset.id = message.message_id; // 存储消息ID
    
    // 添加额外类名，用于样式控制
    if (message.type === MESSAGE_TYPES.IMAGE) {
        messageElement.classList.add('image-message');
    } else if (message.type === MESSAGE_TYPES.EMOJI) {
        messageElement.classList.add('emoji-message');
    } else if (message.type === MESSAGE_TYPES.FILE) {
        messageElement.classList.add('file-message');
    }
    
    // 如果消息已撤回
    if (message.status === 1) {
        messageElement.classList.add('recalled-message');
    }
    
    // 消息头部信息（用户名/IP）
    const messageInfo = document.createElement('div');
    messageInfo.className = 'message-info';
    const displayName = message.username || message.ip;
    messageInfo.textContent = isOwnMessage ? '我' : displayName;
    
    // 消息内容
    const messageContent = document.createElement('div');
    messageContent.className = 'message-content';
    
    if (message.status === 1) {
        // 撤回的消息
        messageContent.textContent = '此消息已被撤回';
    } else {
        // 根据消息类型处理内容
        if (message.type === MESSAGE_TYPES.TEXT) {
            messageContent.textContent = message.content;
        } else if (message.type === MESSAGE_TYPES.IMAGE) {
            const img = document.createElement('img');
            img.src = message.content;
            img.alt = 'Image';
            img.loading = 'lazy';
            messageContent.appendChild(img);
        } else if (message.type === MESSAGE_TYPES.EMOJI) {
            messageContent.textContent = message.content;
        } else if (message.type === MESSAGE_TYPES.FILE) {
            // 文件消息
            const fileInfo = document.createElement('div');
            fileInfo.className = 'file-info';
            
            const fileIcon = document.createElement('div');
            fileIcon.className = 'file-icon';
            fileIcon.textContent = '📄';
            
            const fileDetails = document.createElement('div');
            fileDetails.className = 'file-details';
            
            const fileName = document.createElement('div');
            fileName.className = 'file-name';
            fileName.textContent = message.file_name;
            
            const fileSize = document.createElement('div');
            fileSize.className = 'file-size';
            fileSize.textContent = formatFileSize(message.file_size);
            
            const fileDownload = document.createElement('div');
            fileDownload.className = 'file-download';
            fileDownload.textContent = '点击下载';
            fileDownload.onclick = function() {
                downloadFile(message.content, message.file_name);
            };
            
            fileDetails.appendChild(fileName);
            fileDetails.appendChild(fileSize);
            fileDetails.appendChild(fileDownload);
            
            fileInfo.appendChild(fileIcon);
            fileInfo.appendChild(fileDetails);
            
            messageContent.appendChild(fileInfo);
        }
    }
    
    // 消息时间
    const messageTime = document.createElement('div');
    messageTime.className = 'message-time';
    const timestamp = message.created_at ? new Date(message.created_at) : new Date();
    messageTime.textContent = formatTime(timestamp);
    
    // 添加消息操作按钮（仅自己的消息且未被撤回）
    if (isOwnMessage && message.status !== 1) {
        // 检查是否在8小时内
        const messageTime = message.created_at ? new Date(message.created_at) : new Date();
        const now = new Date();
        const hoursDiff = (now - messageTime) / (1000 * 60 * 60);
        
        if (hoursDiff <= 8) {
            const messageActions = document.createElement('div');
            messageActions.className = 'message-actions';
            
            const recallBtn = document.createElement('button');
            recallBtn.className = 'message-action-btn';
            recallBtn.textContent = '撤回';
            recallBtn.onclick = function(e) {
                e.stopPropagation();
                recallMessage(message.message_id);
            };
            
            messageActions.appendChild(recallBtn);
            messageElement.appendChild(messageActions);
        }
    }
    
    // 组装消息
    messageElement.appendChild(messageInfo);
    messageElement.appendChild(messageContent);
    messageElement.appendChild(messageTime);
    messagesContainer.appendChild(messageElement);
    
    // 将消息添加到映射表
    if (message.message_id) {
        messageMap.set(message.message_id, messageElement);
    }
}

// 处理已撤回的消息
function handleRecalledMessage(message) {
    // 查找要撤回的消息
    const messageElement = messageMap.get(message.message_id);
    if (messageElement) {
        // 添加已撤回样式
        messageElement.classList.add('recalled-message');
        
        // 更新消息内容
        const messageContent = messageElement.querySelector('.message-content');
        messageContent.textContent = '此消息已被撤回';
        
        // 移除操作按钮
        const actionButtons = messageElement.querySelector('.message-actions');
        if (actionButtons) {
            actionButtons.remove();
        }
    }
}

// 渲染系统消息
function renderSystemMessage(message) {
    const messageElement = document.createElement('div');
    messageElement.className = 'message system-message';
    
    const messageContent = document.createElement('div');
    messageContent.className = 'message-content';
    messageContent.textContent = message.content;
    
    messageElement.appendChild(messageContent);
    messagesContainer.appendChild(messageElement);
}

// 撤回消息
function recallMessage(messageId) {
    if (!messageId) return;
    
    const message = {
        type: MESSAGE_TYPES.RECALL,
        message_id: messageId
    };
    
    sendMessage(message);
}

// 下载文件
function downloadFile(dataUrl, fileName) {
    const link = document.createElement('a');
    link.href = dataUrl;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

// 初始化表情列表
function initEmojis() {
    const emojis = [
        '😀', '😁', '😂', '🤣', '😃', '😄', '😅', '😆', '😉', '😊', 
        '😋', '😎', '😍', '😘', '🥰', '😗', '😙', '😚', '🙂', '🤗',
        '🤩', '🤔', '🤨', '😐', '😑', '😶', '🙄', '😏', '😣', '😥',
        '😮', '🤐', '😯', '😪', '😫', '🥱', '😴', '😌', '😛', '😜',
        '😝', '🤤', '😒', '😓', '😔', '😕', '🙃', '🤑', '😲', '☹️',
        '🙁', '😖', '😞', '😟', '😤', '😢', '😭', '😦', '😧', '😨',
        '😩', '🤯', '😬', '😰', '😱', '🥵', '🥶', '😳', '🤪', '😵',
        '🥴', '😠', '😡', '🤬', '😷', '🤒', '🤕', '🤢', '🤮', '🤧'
    ];
    
    emojis.forEach(emoji => {
        const emojiElement = document.createElement('div');
        emojiElement.className = 'emoji';
        emojiElement.textContent = emoji;
        emojiElement.addEventListener('click', () => {
            sendEmojiMessage(emoji);
            toggleEmojiPanel(false);
        });
        emojiList.appendChild(emojiElement);
    });
}

// 获取历史消息
function fetchMessages() {
    fetch('/api/messages')
        .then(response => response.json())
        .then(messages => {
            messages.forEach(message => {
                if (message.type === 'system') {
                    renderSystemMessage({
                        type: MESSAGE_TYPES.SYSTEM,
                        content: message.content
                    });
                } else {
                    renderMessage({
                        type: message.type,
                        content: message.content,
                        username: message.username,
                        ip: message.ip,
                        created_at: message.created_at
                    });
                }
            });
            scrollToBottom();
        })
        .catch(error => console.error('获取历史消息失败:', error));
}

// 获取在线用户
function fetchOnlineUsers() {
    fetch('/api/users/online')
        .then(response => response.json())
        .then(users => {
            renderUserList(users);
        })
        .catch(error => console.error('获取在线用户失败:', error));
}

// 渲染用户列表
function renderUserList(users) {
    userList.innerHTML = '';
    
    if (users.length === 0) {
        const noUsers = document.createElement('div');
        noUsers.textContent = '暂无在线用户';
        userList.appendChild(noUsers);
        return;
    }
    
    users.forEach(user => {
        const userElement = document.createElement('div');
        userElement.className = 'user-item';
        
        const userName = document.createElement('div');
        userName.className = 'user-item-name';
        userName.textContent = user.username || '未设置昵称';
        
        const userIp = document.createElement('div');
        userIp.className = 'user-item-ip';
        userIp.textContent = `IP: ${user.ip}`;
        
        const userTime = document.createElement('div');
        userTime.className = 'user-item-time';
        const lastOnline = new Date(user.last_online);
        userTime.textContent = `最后在线: ${formatDateTime(lastOnline)}`;
        
        userElement.appendChild(userName);
        userElement.appendChild(userIp);
        userElement.appendChild(userTime);
        userList.appendChild(userElement);
        
        // 如果是当前用户，保存用户ID
        if (user.ip === currentUserIP) {
            localUserID = user.id;
        }
    });
}

// 获取聊天室统计信息
function fetchStats() {
    fetch('/api/stats')
        .then(response => response.json())
        .then(data => {
            renderStats(data);
        })
        .catch(error => console.error('获取统计信息失败:', error));
}

// 渲染统计信息
function renderStats(data) {
    statsContent.innerHTML = '';
    
    // 用户数量
    const userCount = document.createElement('div');
    userCount.className = 'stat-item';
    userCount.innerHTML = `<div class="stat-title">总用户数</div>${data.user_count || 0}`;
    statsContent.appendChild(userCount);
    
    // 消息数量
    const messageCount = document.createElement('div');
    messageCount.className = 'stat-item';
    messageCount.innerHTML = `<div class="stat-title">总消息数</div>${data.message_count || 0}`;
    statsContent.appendChild(messageCount);
    
    // 在线用户数
    const onlineUsers = data.online_users || [];
    const onlineUserCount = document.createElement('div');
    onlineUserCount.className = 'stat-item';
    onlineUserCount.innerHTML = `<div class="stat-title">在线用户数</div>${data.active_user_count || onlineUsers.length}`;
    statsContent.appendChild(onlineUserCount);
}

// 搜索消息
function searchMessages() {
    const query = searchInput.value.trim();
    if (!query) return;
    
    fetch(`/api/messages/search?q=${encodeURIComponent(query)}`)
        .then(response => response.json())
        .then(messages => {
            renderSearchResults(messages);
        })
        .catch(error => console.error('搜索消息失败:', error));
}

// 渲染搜索结果
function renderSearchResults(messages) {
    searchResults.innerHTML = '';
    
    if (messages.length === 0) {
        const noResults = document.createElement('div');
        noResults.textContent = '未找到匹配的消息';
        searchResults.appendChild(noResults);
        return;
    }
    
    messages.forEach(message => {
        const searchItem = document.createElement('div');
        searchItem.className = 'search-item';
        
        const searchItemInfo = document.createElement('div');
        searchItemInfo.className = 'search-item-info';
        
        const userInfo = document.createElement('span');
        userInfo.textContent = `${message.username || message.ip}`;
        
        const timeInfo = document.createElement('span');
        const msgTime = new Date(message.created_at);
        timeInfo.textContent = formatDateTime(msgTime);
        
        searchItemInfo.appendChild(userInfo);
        searchItemInfo.appendChild(timeInfo);
        
        const contentPreview = document.createElement('div');
        if (message.type === 'text') {
            contentPreview.textContent = message.content;
        } else if (message.type === 'image') {
            contentPreview.textContent = '[图片消息]';
        } else if (message.type === 'emoji') {
            contentPreview.textContent = '[表情消息]';
        } else {
            contentPreview.textContent = message.content;
        }
        
        searchItem.appendChild(searchItemInfo);
        searchItem.appendChild(contentPreview);
        searchResults.appendChild(searchItem);
    });
}

// 绑定事件
function bindEvents() {
    // 发送消息
    sendButton.addEventListener('click', sendTextMessage);
    messageInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendTextMessage();
        }
    });
    
    // 设置用户名
    setUsernameButton.addEventListener('click', setUsername);
    
    // 上传图片
    imageUpload.addEventListener('change', handleImageUpload);
    
    // 上传文件
    fileUpload.addEventListener('change', handleFileUpload);
    
    // 表情面板
    emojiToggle.addEventListener('click', () => toggleEmojiPanel());
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.emoji-picker') && emojiPanel.style.display === 'block') {
            toggleEmojiPanel(false);
        }
    });
    
    // 搜索消息
    searchButton.addEventListener('click', searchMessages);
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            searchMessages();
        }
    });
    
    // 标签切换
    tabButtons.forEach(button => {
        button.addEventListener('click', () => {
            const tabName = button.getAttribute('data-tab');
            switchTab(tabName);
        });
    });
}

// 发送文本消息
function sendTextMessage() {
    const content = messageInput.value.trim();
    if (!content) return;
    
    const message = {
        type: MESSAGE_TYPES.TEXT,
        content
    };
    
    sendMessage(message);
    messageInput.value = '';
}

// 发送表情消息
function sendEmojiMessage(emoji) {
    const message = {
        type: MESSAGE_TYPES.EMOJI,
        content: emoji
    };
    
    sendMessage(message);
}

// 处理图片上传
function handleImageUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    
    if (!file.type.match('image.*')) {
        alert('请选择图片文件');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        const message = {
            type: MESSAGE_TYPES.IMAGE,
            content: e.target.result
        };
        
        sendMessage(message);
    };
    
    reader.readAsDataURL(file);
    // 清空文件选择器，使同一文件可以再次选择
    imageUpload.value = '';
}

// 处理文件上传
function handleFileUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    
    // 文件大小限制（10MB）
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
        alert('文件大小不能超过10MB');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        const message = {
            type: MESSAGE_TYPES.FILE,
            content: e.target.result,
            file_name: file.name,
            file_size: file.size
        };
        
        sendMessage(message);
    };
    
    reader.readAsDataURL(file);
    // 清空文件选择器，使同一文件可以再次选择
    fileUpload.value = '';
}

// 设置用户名
function setUsername() {
    const username = usernameInput.value.trim();
    if (!username) return;
    
    const message = {
        type: MESSAGE_TYPES.USER,
        username
    };
    
    sendMessage(message);
}

// 发送消息到服务器
function sendMessage(message) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(message));
    } else {
        console.error('WebSocket未连接');
    }
}

// 切换表情面板
function toggleEmojiPanel(show) {
    if (show === undefined) {
        emojiPanel.style.display = emojiPanel.style.display === 'block' ? 'none' : 'block';
    } else {
        emojiPanel.style.display = show ? 'block' : 'none';
    }
}

// 切换标签页
function switchTab(tabName) {
    tabButtons.forEach(btn => {
        if (btn.getAttribute('data-tab') === tabName) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });
    
    tabContents.forEach(content => {
        if (content.id === tabName) {
            content.classList.add('active');
        } else {
            content.classList.remove('active');
        }
    });
    
    // 如果切换到统计标签，刷新统计信息
    if (tabName === 'statistics') {
        fetchStats();
    } else if (tabName === 'online-users') {
        fetchOnlineUsers();
    }
}

// 滚动到底部
function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// 格式化时间，显示时分
function formatTime(date) {
    return `${padZero(date.getHours())}:${padZero(date.getMinutes())}`;
}

// 格式化日期时间，显示年月日时分
function formatDateTime(date) {
    return `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())} ${padZero(date.getHours())}:${padZero(date.getMinutes())}`;
}

// 补零
function padZero(num) {
    return num < 10 ? `0${num}` : num;
}

// 页面加载时初始化
window.addEventListener('load', init); 