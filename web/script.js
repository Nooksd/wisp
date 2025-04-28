// Dados do usuário atual
const currentUser = {
  userId: "joao321",
  name: "Usuário Web",
  email: "teste@gmail.com",
};

// Lista de contatos (simulada)
const contacts = [
  {
    userId: "joao123",
    name: "Usuário Flutter",
    email: "teste2@gmail.com",
  },
];

// Estado da aplicação
let selectedContact = null;
let socket = null;
const messages = {};

// Elementos DOM
const contactsPage = document.getElementById("contacts-page");
const chatPage = document.getElementById("chat-page");
const contactList = document.getElementById("contact-list");
const messagesContainer = document.getElementById("messages");
const messageInput = document.getElementById("message-input");
const sendButton = document.getElementById("send-button");
const chatTitle = document.getElementById("chat-title");

// Renderizar lista de contatos
function renderContacts() {
  contactList.innerHTML = "";
  contacts.forEach((contact) => {
    const contactItem = document.createElement("div");
    contactItem.className = "contact-item";
    contactItem.innerHTML = `
                <h3>${contact.name}</h3>
                <p>${contact.email}</p>
            `;
    contactItem.onclick = () => openChat(contact);
    contactList.appendChild(contactItem);
  });
}

// Abrir chat com um contato
function openChat(contact) {
  selectedContact = contact;
  chatTitle.textContent = contact.name;

  // Limpar mensagens exibidas e mostrar as do contato selecionado
  messagesContainer.innerHTML = "";

  if (messages[contact.userId]) {
    messages[contact.userId].forEach((msg) => {
      appendMessage(msg.content, msg.sent);
    });
  }

  contactsPage.style.display = "none";
  chatPage.style.display = "block";
}

// Voltar para a lista de contatos
function showContactsPage() {
  chatPage.style.display = "none";
  contactsPage.style.display = "block";
}

// Adicionar mensagem ao chat
function appendMessage(content, isSent) {
  const messageDiv = document.createElement("div");
  messageDiv.className = `message ${isSent ? "sent" : "received"}`;
  messageDiv.textContent = content;
  messagesContainer.appendChild(messageDiv);
  messagesContainer.scrollTop = messagesContainer.scrollHeight;

  // Salvar a mensagem no estado
  if (selectedContact) {
    if (!messages[selectedContact.userId]) {
      messages[selectedContact.userId] = [];
    }
    messages[selectedContact.userId].push({
      content,
      sent: isSent,
      timestamp: new Date(),
    });
  }
}

// Enviar mensagem
function sendMessage() {
  const content = messageInput.value.trim();
  if (!content || !selectedContact) return;

  const message = {
    type: "message",
    from: currentUser.userId,
    to: selectedContact.userId,
    content: content,
    msgId: generateMsgId(),
  };

  // Enviar via WebSocket
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
    appendMessage(content, true);
    messageInput.value = "";
  } else {
    console.error("WebSocket não está conectado");
    reconnectWebSocket();
  }
}

// Gerar ID único para mensagem
function generateMsgId() {
  return Date.now().toString() + Math.random().toString(36).substr(2, 9);
}

// Configurar WebSocket
function setupWebSocket() {
  const token =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJqb2FvMzIxIiwiaXNBZG1pbiI6dHJ1ZSwiZXhwIjoxNzQ2NDY5NTg0LCJpYXQiOjE3NDU4NjQ3ODQsImp0aSI6IjY4MGZjODUwM2IwYjU5ZDk3OTBjMjFmYSJ9.ORDClLhH4tZwSuyoHGyh0uupv6oz-osC8NVwOBrumuw";
  const deviceId = "33123754";

  const wsUrl = `ws://192.168.1.20:9999/ws?deviceId=33123754&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJqb2FvMzIxIiwiaXNBZG1pbiI6dHJ1ZSwiZXhwIjoxNzQ2NDY5NTg0LCJpYXQiOjE3NDU4NjQ3ODQsImp0aSI6IjY4MGZjODUwM2IwYjU5ZDk3OTBjMjFmYSJ9.ORDClLhH4tZwSuyoHGyh0uupv6oz-osC8NVwOBrumuw`;
  socket = new WebSocket(wsUrl);

  socket.onopen = function () {
    console.log("Conectado ao WebSocket");
    socket.send(
      JSON.stringify({
        type: "auth",
        token: token,
      })
    );
  };

  socket.onmessage = function (event) {
    try {
      const data = JSON.parse(event.data);
      console.log("Mensagem recebida:", data);

      if (data.type === "message") {
        if (selectedContact && data.from === selectedContact.userId) {
          appendMessage(data.content, false);
        } else {
          if (!messages[data.from]) {
            messages[data.from] = [];
          }
          messages[data.from].push({
            content: data.content,
            sent: false,
            timestamp: new Date(),
          });
        }

        socket.send(
          JSON.stringify({
            type: "receipt",
            from: currentUser.userId,
            to: data.from,
            msgId: data.msgId,
            status: "delivered",
          })
        );
      } else if (data.type === "receipt") {
        console.log("Recibo recebido:", data);
        // Poderia atualizar UI para mostrar status da mensagem
      }
    } catch (e) {
      console.error("Erro ao processar mensagem:", e);
    }
  };

  socket.onclose = function () {
    console.log("Conexão WebSocket fechada");
    setTimeout(reconnectWebSocket, 3000);
  };

  socket.onerror = function (error) {
    console.error("Erro WebSocket:", error);
  };
}

function reconnectWebSocket() {
  console.log("Tentando reconectar WebSocket...");
  setupWebSocket();
}

// Event listeners
sendButton.addEventListener("click", sendMessage);
messageInput.addEventListener("keypress", function (e) {
  if (e.key === "Enter") {
    sendMessage();
  }
});

// Inicialização
document.addEventListener("DOMContentLoaded", function () {
  renderContacts();
  setupWebSocket();
});
