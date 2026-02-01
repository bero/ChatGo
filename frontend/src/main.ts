// Main entry point for ChatGO frontend

const API_URL = "http://localhost:8080";
const WS_URL = "ws://localhost:8080";

// DOM elements - Login
const loginSection = document.getElementById("login-section") as HTMLDivElement;
const loginForm = document.getElementById("login-form") as HTMLFormElement;
const loginMessage = document.getElementById("login-message") as HTMLDivElement;

// DOM elements - Chat
const chatSection = document.getElementById("chat-section") as HTMLDivElement;
const currentUserDisplay = document.getElementById("current-user") as HTMLDivElement;
const conversationsList = document.getElementById("conversations-list") as HTMLDivElement;
const usersList = document.getElementById("users-list") as HTMLDivElement;
const newGroupBtn = document.getElementById("new-group-btn") as HTMLButtonElement;
const chatPlaceholder = document.getElementById("chat-placeholder") as HTMLDivElement;
const activeChat = document.getElementById("active-chat") as HTMLDivElement;
const chatWith = document.getElementById("chat-with") as HTMLSpanElement;
const typingIndicator = document.getElementById("typing-indicator") as HTMLSpanElement;
const messagesContainer = document.getElementById("messages") as HTMLDivElement;
const messageInput = document.getElementById("message-input") as HTMLInputElement;
const sendBtn = document.getElementById("send-btn") as HTMLButtonElement;
const logoutBtn = document.getElementById("logout-btn") as HTMLButtonElement;

// DOM elements - Group Modal
const groupModal = document.getElementById("group-modal") as HTMLDivElement;
const createGroupForm = document.getElementById("create-group-form") as HTMLFormElement;
const groupNameInput = document.getElementById("group-name") as HTMLInputElement;
const groupUserList = document.getElementById("group-user-list") as HTMLDivElement;
const cancelGroupBtn = document.getElementById("cancel-group-btn") as HTMLButtonElement;
const groupMessage = document.getElementById("group-message") as HTMLDivElement;

// DOM elements - Admin
const adminBtn = document.getElementById("admin-btn") as HTMLButtonElement;
const adminPanel = document.getElementById("admin-panel") as HTMLDivElement;
const adminUserList = document.getElementById("admin-user-list") as HTMLDivElement;
const backToChatBtn = document.getElementById("back-to-chat-btn") as HTMLButtonElement;
const createUserForm = document.getElementById("create-user-form") as HTMLFormElement;
const createUserMessage = document.getElementById("create-user-message") as HTMLDivElement;

// DOM elements - Edit Modal
const editModal = document.getElementById("edit-modal") as HTMLDivElement;
const editUserForm = document.getElementById("edit-user-form") as HTMLFormElement;
const editUserId = document.getElementById("edit-user-id") as HTMLInputElement;
const editUsername = document.getElementById("edit-username") as HTMLInputElement;
const editPassword = document.getElementById("edit-password") as HTMLInputElement;
const editIsAdmin = document.getElementById("edit-is-admin") as HTMLInputElement;
const cancelEditBtn = document.getElementById("cancel-edit-btn") as HTMLButtonElement;
const editUserMessage = document.getElementById("edit-user-message") as HTMLDivElement;

// State
let authToken: string | null = null;
let currentUserId: string | null = null;
let currentUsername: string | null = null;
let currentUserIsAdmin: boolean = false;
let selectedUserId: string | null = null;
let currentConversationId: string | null = null;
let currentConversationName: string | null = null;
let websocket: WebSocket | null = null;
let typingTimeout: number | null = null;
let allUsers: User[] = [];
let unreadCounts: Map<string, number> = new Map(); // conversationId -> unread count

// User interface
interface User {
    id: string;
    username: string;
    is_admin: boolean;
}

// Participant interface
interface Participant {
    id: string;
    username: string;
}

// Conversation interface
interface Conversation {
    id: string;
    name?: string;
    is_group: boolean;
    participants: Participant[];
    created_at: string;
}

// Message interface
interface ChatMessage {
    type: string;
    id: string;
    conversation_id: string;
    sender_id: string;
    sender_username: string;
    content: string;
    created_at: string;
}

// Typing message interface
interface TypingMessage {
    type: string;
    conversation_id: string;
    user_id: string;
    username: string;
    is_typing: boolean;
}

// Initialize the app
function init(): void {
    const savedToken = localStorage.getItem("token");
    const savedUsername = localStorage.getItem("username");
    const savedUserId = localStorage.getItem("userId");
    const savedIsAdmin = localStorage.getItem("isAdmin");

    if (savedToken && savedUsername && savedUserId) {
        authToken = savedToken;
        currentUsername = savedUsername;
        currentUserId = savedUserId;
        currentUserIsAdmin = savedIsAdmin === "true";
        showChatSection();
    } else {
        showLoginSection();
    }

    // Event listeners - Login/Chat
    loginForm.addEventListener("submit", handleLogin);
    logoutBtn.addEventListener("click", handleLogout);
    sendBtn.addEventListener("click", sendMessage);
    messageInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
            sendMessage();
        }
    });
    messageInput.addEventListener("input", handleTyping);

    // Event listeners - Admin
    adminBtn.addEventListener("click", showAdminPanel);
    backToChatBtn.addEventListener("click", hideAdminPanel);
    createUserForm.addEventListener("submit", handleCreateUser);
    editUserForm.addEventListener("submit", handleEditUser);
    cancelEditBtn.addEventListener("click", closeEditModal);

    // Event listeners - Group
    newGroupBtn.addEventListener("click", showGroupModal);
    createGroupForm.addEventListener("submit", handleCreateGroup);
    cancelGroupBtn.addEventListener("click", closeGroupModal);
}

function showLoginSection(): void {
    loginSection.style.display = "block";
    chatSection.style.display = "none";
}

function showChatSection(): void {
    loginSection.style.display = "none";
    chatSection.style.display = "flex";
    currentUserDisplay.textContent = `Logged in as: ${currentUsername}`;

    // Show admin button only for admins.
    adminBtn.style.display = currentUserIsAdmin ? "block" : "none";

    loadUsersAndConversations();
    connectWebSocket();
}

// Handle login
async function handleLogin(event: Event): Promise<void> {
    event.preventDefault();

    const usernameInput = document.getElementById("username") as HTMLInputElement;
    const passwordInput = document.getElementById("password") as HTMLInputElement;

    try {
        const response = await fetch(`${API_URL}/api/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                username: usernameInput.value,
                password: passwordInput.value
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            showLoginMessage(data.error || "Login failed", "error");
            return;
        }

        // Save auth info
        authToken = data.token;
        currentUsername = data.username;

        // Decode JWT to get user ID and admin status
        const payload = JSON.parse(atob(data.token.split('.')[1]));
        currentUserId = payload.user_id;
        currentUserIsAdmin = payload.is_admin === true;

        localStorage.setItem("token", data.token);
        localStorage.setItem("username", data.username);
        localStorage.setItem("userId", currentUserId!);
        localStorage.setItem("isAdmin", currentUserIsAdmin.toString());

        showChatSection();

    } catch (error) {
        showLoginMessage("Failed to connect to server", "error");
        console.error("Login error:", error);
    }
}

function handleLogout(): void {
    authToken = null;
    currentUserId = null;
    currentUsername = null;
    currentUserIsAdmin = false;
    selectedUserId = null;
    currentConversationId = null;

    if (websocket) {
        websocket.close();
        websocket = null;
    }

    localStorage.removeItem("token");
    localStorage.removeItem("username");
    localStorage.removeItem("userId");
    localStorage.removeItem("isAdmin");

    showLoginSection();
}

function showLoginMessage(message: string, type: "error" | "success"): void {
    loginMessage.textContent = message;
    loginMessage.className = type;
}

// Load users and conversations for the sidebar
async function loadUsersAndConversations(): Promise<void> {
    try {
        // Load users
        const usersResponse = await fetch(`${API_URL}/api/users`, {
            headers: { "Authorization": `Bearer ${authToken}` }
        });

        if (!usersResponse.ok) {
            console.error("Failed to load users");
            return;
        }

        allUsers = await usersResponse.json();

        // Load conversations
        const convsResponse = await fetch(`${API_URL}/api/conversations`, {
            headers: { "Authorization": `Bearer ${authToken}` }
        });

        let conversations: Conversation[] = [];
        if (convsResponse.ok) {
            conversations = await convsResponse.json() || [];
        }

        // Clear and populate conversations list
        conversationsList.innerHTML = "";

        // Show group conversations first
        conversations
            .filter(conv => conv.is_group)
            .forEach(conv => {
                const convItem = document.createElement("div");
                convItem.className = "conversation-item group-chat";
                convItem.dataset.conversationId = conv.id;
                const unreadCount = unreadCounts.get(conv.id) || 0;
                const badgeDisplay = unreadCount > 0 ? "inline" : "none";
                convItem.innerHTML = `
                    <div class="conv-name">${escapeHtml(conv.name || "Group")} <span class="unread-badge" style="display: ${badgeDisplay}">(${unreadCount})</span></div>
                    <div class="conv-info">${conv.participants.length} members</div>
                `;
                convItem.addEventListener("click", () => selectConversation(conv));
                conversationsList.appendChild(convItem);
            });

        // Clear and populate users list (for direct messages)
        usersList.innerHTML = "";

        // Get IDs of users we already have 1:1 conversations with
        const usersWithConversations = new Set<string>();
        conversations
            .filter(conv => !conv.is_group && conv.participants.length === 2)
            .forEach(conv => {
                const otherUser = conv.participants.find(p => p.id !== currentUserId);
                if (otherUser) {
                    usersWithConversations.add(otherUser.id);
                    // Add existing 1:1 conversation to the list
                    const convItem = document.createElement("div");
                    convItem.className = "conversation-item";
                    convItem.dataset.conversationId = conv.id;
                    convItem.dataset.userId = otherUser.id;
                    const unreadCount = unreadCounts.get(conv.id) || 0;
                    const badgeDisplay = unreadCount > 0 ? "inline" : "none";
                    convItem.innerHTML = `
                        <div class="conv-name">${escapeHtml(otherUser.username)} <span class="unread-badge" style="display: ${badgeDisplay}">(${unreadCount})</span></div>
                        <div class="conv-info">Direct message</div>
                    `;
                    convItem.addEventListener("click", () => selectConversation(conv));
                    usersList.appendChild(convItem);
                }
            });

        // Add users we don't have conversations with
        allUsers
            .filter(user => user.id !== currentUserId && !usersWithConversations.has(user.id))
            .forEach(user => {
                const userItem = document.createElement("div");
                userItem.className = "user-item";
                userItem.dataset.userId = user.id;
                userItem.dataset.username = user.username;
                userItem.innerHTML = `
                    <div class="username">${escapeHtml(user.username)}</div>
                    <div class="status">${user.is_admin ? "Admin" : "User"}</div>
                `;
                userItem.addEventListener("click", () => selectUser(user.id, user.username));
                usersList.appendChild(userItem);
            });

    } catch (error) {
        console.error("Error loading users/conversations:", error);
    }
}

// Legacy function for compatibility
async function loadUsers(): Promise<void> {
    await loadUsersAndConversations();
}

// Select a user to chat with (creates 1:1 conversation)
async function selectUser(userId: string, username: string): Promise<void> {
    selectedUserId = userId;
        currentConversationName = username;

    // Update UI - clear all active states
    document.querySelectorAll(".user-item, .conversation-item").forEach(item => {
        item.classList.remove("active");
    });
    document.querySelectorAll(".user-item").forEach(item => {
        if ((item as HTMLElement).dataset.userId === userId) {
            item.classList.add("active");
        }
    });

    chatPlaceholder.style.display = "none";
    activeChat.style.display = "flex";
    chatWith.textContent = username;
    typingIndicator.textContent = "";
    messagesContainer.innerHTML = "";

    // Get or create conversation
    await getOrCreateConversation(userId);
}

// Select an existing conversation (1:1 or group)
async function selectConversation(conv: Conversation): Promise<void> {
    currentConversationId = conv.id;

    // Clear unread count for this conversation
    unreadCounts.delete(conv.id);
    updateUnreadBadges();

    // Update UI - clear all active states
    document.querySelectorAll(".user-item, .conversation-item").forEach(item => {
        item.classList.remove("active");
    });
    document.querySelectorAll(".conversation-item").forEach(item => {
        if ((item as HTMLElement).dataset.conversationId === conv.id) {
            item.classList.add("active");
        }
    });

    chatPlaceholder.style.display = "none";
    activeChat.style.display = "flex";

    if (conv.is_group) {
        currentConversationName = conv.name || "Group";
        chatWith.textContent = `${conv.name} (${conv.participants.length} members)`;
    } else {
        const otherUser = conv.participants.find(p => p.id !== currentUserId);
        selectedUserId = otherUser?.id || null;
                currentConversationName = otherUser?.username || "Chat";
        chatWith.textContent = currentConversationName;
    }

    typingIndicator.textContent = "";
    messagesContainer.innerHTML = "";

    // Load messages
    await loadMessages(conv.id);
}

// Get or create a conversation with another user
async function getOrCreateConversation(otherUserId: string): Promise<void> {
    try {
        const response = await fetch(`${API_URL}/api/conversations`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify({ other_user_id: otherUserId })
        });

        if (!response.ok) {
            console.error("Failed to get/create conversation");
            return;
        }

        const data = await response.json();
        currentConversationId = data.id;

        // Load messages for this conversation
        await loadMessages(data.id);

    } catch (error) {
        console.error("Error getting conversation:", error);
    }
}

// Load messages for a conversation
async function loadMessages(conversationId: string): Promise<void> {
    try {
        const response = await fetch(`${API_URL}/api/conversations/${conversationId}/messages`, {
            headers: { "Authorization": `Bearer ${authToken}` }
        });

        if (!response.ok) {
            console.error("Failed to load messages");
            return;
        }

        const messages: ChatMessage[] = await response.json();

        messagesContainer.innerHTML = "";

        if (messages) {
            messages.forEach(msg => addMessageToUI(msg));
        }

        // Scroll to bottom
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

    } catch (error) {
        console.error("Error loading messages:", error);
    }
}

// Connect to WebSocket
function connectWebSocket(): void {
    if (websocket) {
        websocket.close();
    }

    websocket = new WebSocket(`${WS_URL}/ws?token=${authToken}`);

    websocket.onopen = (): void => {
        console.log("WebSocket connected");
    };

    websocket.onmessage = (event): void => {
        const data = JSON.parse(event.data);

        if (data.type === "message") {
            handleIncomingMessage(data as ChatMessage);
        } else if (data.type === "typing") {
            handleTypingIndicator(data as TypingMessage);
        } else if (data.type === "new_conversation") {
            // Refresh conversation list when added to a new conversation
            loadUsersAndConversations();
        }
    };

    websocket.onclose = (): void => {
        console.log("WebSocket disconnected");
        // Reconnect after 3 seconds
        setTimeout(() => {
            if (authToken) {
                connectWebSocket();
            }
        }, 3000);
    };

    websocket.onerror = (error): void => {
        console.error("WebSocket error:", error);
    };
}

// Handle incoming chat message
function handleIncomingMessage(msg: ChatMessage): void {
    if (msg.conversation_id === currentConversationId) {
        // Currently viewing this conversation - show message
        addMessageToUI(msg);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    } else {
        // Not viewing this conversation - increment unread count (only for messages from others)
        if (msg.sender_id !== currentUserId) {
            const currentCount = unreadCounts.get(msg.conversation_id) || 0;
            unreadCounts.set(msg.conversation_id, currentCount + 1);
            updateUnreadBadges();
        }
    }
}

// Update unread badges in the sidebar
function updateUnreadBadges(): void {
    // Update conversation items
    document.querySelectorAll(".conversation-item").forEach(item => {
        const convId = (item as HTMLElement).dataset.conversationId;
        const badge = item.querySelector(".unread-badge") as HTMLElement;
        if (convId && badge) {
            const count = unreadCounts.get(convId) || 0;
            if (count > 0) {
                badge.textContent = `(${count})`;
                badge.style.display = "inline";
            } else {
                badge.style.display = "none";
            }
        }
    });

    // Update user items (for users without existing conversations)
    document.querySelectorAll(".user-item").forEach(item => {
        const convId = (item as HTMLElement).dataset.conversationId;
        const badge = item.querySelector(".unread-badge") as HTMLElement;
        if (convId && badge) {
            const count = unreadCounts.get(convId) || 0;
            if (count > 0) {
                badge.textContent = `(${count})`;
                badge.style.display = "inline";
            } else {
                badge.style.display = "none";
            }
        }
    });
}

// Handle typing indicator
function handleTypingIndicator(data: TypingMessage): void {
    if (data.conversation_id === currentConversationId && data.user_id !== currentUserId) {
        if (data.is_typing) {
            typingIndicator.textContent = ` - ${data.username} is typing...`;
        } else {
            typingIndicator.textContent = "";
        }
    }
}

// Add a message to the UI
function addMessageToUI(msg: ChatMessage): void {
    const messageDiv = document.createElement("div");
    const isSent = msg.sender_id === currentUserId;
    messageDiv.className = `message ${isSent ? "sent" : "received"}`;

    const time = new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

    messageDiv.innerHTML = `
        <div class="content">${escapeHtml(msg.content)}</div>
        <div class="time">${time}</div>
    `;

    messagesContainer.appendChild(messageDiv);
}

// Send a message
function sendMessage(): void {
    const content = messageInput.value.trim();

    if (!content || !websocket || !currentConversationId) {
        return;
    }

    // Send via WebSocket
    websocket.send(JSON.stringify({
        type: "message",
        conversation_id: currentConversationId,
        content: content
    }));

    messageInput.value = "";

    // Send typing stopped
    sendTypingStatus(false);
}

// Handle typing - send typing indicator
function handleTyping(): void {
    if (!websocket || !currentConversationId) {
        return;
    }

    // Send typing started
    sendTypingStatus(true);

    // Clear previous timeout
    if (typingTimeout) {
        clearTimeout(typingTimeout);
    }

    // Send typing stopped after 2 seconds of no input
    typingTimeout = window.setTimeout(() => {
        sendTypingStatus(false);
    }, 2000);
}

function sendTypingStatus(isTyping: boolean): void {
    if (!websocket || !currentConversationId) {
        return;
    }

    websocket.send(JSON.stringify({
        type: "typing",
        conversation_id: currentConversationId,
        is_typing: isTyping
    }));
}

// Escape HTML to prevent XSS
function escapeHtml(text: string): string {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

// ==================== Admin Panel Functions ====================

// Show the admin panel
function showAdminPanel(): void {
    chatPlaceholder.style.display = "none";
    activeChat.style.display = "none";
    adminPanel.style.display = "block";
    loadAdminUsers();
}

// Hide the admin panel and return to chat
function hideAdminPanel(): void {
    adminPanel.style.display = "none";

    // Show chat placeholder or active chat
    if (selectedUserId) {
        activeChat.style.display = "flex";
    } else {
        chatPlaceholder.style.display = "flex";
    }
}

// Load all users for admin management
async function loadAdminUsers(): Promise<void> {
    try {
        const response = await fetch(`${API_URL}/api/users`, {
            headers: { "Authorization": `Bearer ${authToken}` }
        });

        if (!response.ok) {
            console.error("Failed to load users for admin");
            return;
        }

        const users: User[] = await response.json();

        adminUserList.innerHTML = "";

        users.forEach(user => {
            const userItem = document.createElement("div");
            userItem.className = "admin-user-item";

            // Check if this is the current user (can't delete/edit self)
            const isSelf = user.id === currentUserId;

            userItem.innerHTML = `
                <div class="user-info">
                    <div class="name">${escapeHtml(user.username)}${isSelf ? " (you)" : ""}</div>
                    <div class="role">${user.is_admin ? "Administrator" : "User"}</div>
                </div>
                <div class="actions">
                    ${!isSelf ? `
                        <button class="edit-btn" data-id="${user.id}" data-username="${escapeHtml(user.username)}" data-admin="${user.is_admin}">Edit</button>
                        <button class="delete-btn danger" data-id="${user.id}" data-username="${escapeHtml(user.username)}">Delete</button>
                    ` : ""}
                </div>
            `;

            adminUserList.appendChild(userItem);
        });

        // Add event listeners to edit buttons
        document.querySelectorAll(".edit-btn").forEach(btn => {
            btn.addEventListener("click", (e) => {
                const target = e.target as HTMLButtonElement;
                openEditModal(
                    target.dataset.id!,
                    target.dataset.username!,
                    target.dataset.admin === "true"
                );
            });
        });

        // Add event listeners to delete buttons
        document.querySelectorAll(".delete-btn").forEach(btn => {
            btn.addEventListener("click", (e) => {
                const target = e.target as HTMLButtonElement;
                handleDeleteUser(target.dataset.id!, target.dataset.username!);
            });
        });

    } catch (error) {
        console.error("Error loading admin users:", error);
    }
}

// Handle create user form submission
async function handleCreateUser(event: Event): Promise<void> {
    event.preventDefault();

    const usernameInput = document.getElementById("new-username") as HTMLInputElement;
    const passwordInput = document.getElementById("new-password") as HTMLInputElement;
    const isAdminInput = document.getElementById("new-is-admin") as HTMLInputElement;

    try {
        const response = await fetch(`${API_URL}/api/users`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify({
                username: usernameInput.value,
                password: passwordInput.value,
                is_admin: isAdminInput.checked
            })
        });

        const data = await response.json();

        if (!response.ok) {
            showCreateUserMessage(data.error || "Failed to create user", "error");
            return;
        }

        showCreateUserMessage("User created successfully", "success");

        // Clear form
        usernameInput.value = "";
        passwordInput.value = "";
        isAdminInput.checked = false;

        // Reload user lists
        loadAdminUsers();
        loadUsers();

    } catch (error) {
        showCreateUserMessage("Failed to create user", "error");
        console.error("Create user error:", error);
    }
}

function showCreateUserMessage(message: string, type: "error" | "success"): void {
    createUserMessage.textContent = message;
    createUserMessage.className = type;

    // Clear message after 3 seconds
    setTimeout(() => {
        createUserMessage.textContent = "";
        createUserMessage.className = "";
    }, 3000);
}

// Open edit user modal
function openEditModal(userId: string, username: string, isAdmin: boolean): void {
    editUserId.value = userId;
    editUsername.value = username;
    editPassword.value = "";
    editIsAdmin.checked = isAdmin;
    editUserMessage.textContent = "";
    editModal.style.display = "flex";
}

// Close edit user modal
function closeEditModal(): void {
    editModal.style.display = "none";
}

// Handle edit user form submission
async function handleEditUser(event: Event): Promise<void> {
    event.preventDefault();

    const userId = editUserId.value;

    try {
        const response = await fetch(`${API_URL}/api/users/${userId}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify({
                username: editUsername.value,
                password: editPassword.value, // Empty string = don't change
                is_admin: editIsAdmin.checked
            })
        });

        const data = await response.json();

        if (!response.ok) {
            showEditUserMessage(data.error || "Failed to update user", "error");
            return;
        }

        // Close modal and reload
        closeEditModal();
        loadAdminUsers();
        loadUsers();

    } catch (error) {
        showEditUserMessage("Failed to update user", "error");
        console.error("Edit user error:", error);
    }
}

function showEditUserMessage(message: string, type: "error" | "success"): void {
    editUserMessage.textContent = message;
    editUserMessage.className = type;
}

// Handle delete user
async function handleDeleteUser(userId: string, username: string): Promise<void> {
    // Confirm before deleting
    if (!confirm(`Are you sure you want to delete user "${username}"?`)) {
        return;
    }

    try {
        const response = await fetch(`${API_URL}/api/users/${userId}`, {
            method: "DELETE",
            headers: {
                "Authorization": `Bearer ${authToken}`
            }
        });

        if (!response.ok) {
            const data = await response.json();
            alert(data.error || "Failed to delete user");
            return;
        }

        // Reload user lists
        loadAdminUsers();
        loadUsers();

    } catch (error) {
        alert("Failed to delete user");
        console.error("Delete user error:", error);
    }
}

// ==================== Group Chat Functions ====================

// Show the group creation modal
function showGroupModal(): void {
    // Populate user list
    groupUserList.innerHTML = "";

    allUsers
        .filter(user => user.id !== currentUserId)
        .forEach(user => {
            const userItem = document.createElement("div");
            userItem.className = "user-checkbox-item";
            userItem.innerHTML = `
                <input type="checkbox" id="group-user-${user.id}" value="${user.id}">
                <label for="group-user-${user.id}">${escapeHtml(user.username)}</label>
            `;
            groupUserList.appendChild(userItem);
        });

    groupNameInput.value = "";
    groupMessage.textContent = "";
    groupModal.style.display = "flex";
}

// Close the group creation modal
function closeGroupModal(): void {
    groupModal.style.display = "none";
}

// Handle group creation form submission
async function handleCreateGroup(event: Event): Promise<void> {
    event.preventDefault();

    const name = groupNameInput.value.trim();
    if (!name) {
        showGroupMessage("Group name is required", "error");
        return;
    }

    // Get selected users
    const selectedUsers: string[] = [];
    groupUserList.querySelectorAll("input[type='checkbox']:checked").forEach((checkbox) => {
        selectedUsers.push((checkbox as HTMLInputElement).value);
    });

    if (selectedUsers.length < 1) {
        showGroupMessage("Select at least one other user", "error");
        return;
    }

    try {
        const response = await fetch(`${API_URL}/api/conversations`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify({
                name: name,
                participant_ids: selectedUsers
            })
        });

        if (!response.ok) {
            const data = await response.json();
            showGroupMessage(data.error || "Failed to create group", "error");
            return;
        }

        const conv = await response.json();

        // Close modal and refresh
        closeGroupModal();
        await loadUsersAndConversations();

        // Load fresh conversation data and select it
        const convsResponse = await fetch(`${API_URL}/api/conversations`, {
            headers: { "Authorization": `Bearer ${authToken}` }
        });
        if (convsResponse.ok) {
            const conversations: Conversation[] = await convsResponse.json() || [];
            const newConv = conversations.find(c => c.id === conv.id);
            if (newConv) {
                selectConversation(newConv);
            }
        }

    } catch (error) {
        showGroupMessage("Failed to create group", "error");
        console.error("Create group error:", error);
    }
}

function showGroupMessage(message: string, type: "error" | "success"): void {
    groupMessage.textContent = message;
    groupMessage.className = type;
}

// Start the app
init();
