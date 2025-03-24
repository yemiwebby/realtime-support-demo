let client;
let conversation;

async function initChat(identity) {
  console.log(`🚀 Starting chat initialization for ${identity}`);

  try {
    // 1. Fetch the JWT token from backend
    const res = await fetch(`/token?identity=${identity}`);
    if (!res.ok) throw new Error(`❌ Failed to fetch token: ${res.statusText}`);
    const { token } = await res.json();

    console.log(`✅ Token fetched for ${identity}`);
    console.log(`🔑 Token preview: ${token.substring(0, 30)}... (${token.length} chars)`);

    // 2. Initialize Twilio Conversations client
    client = new Twilio.Conversations.Client(token);
    console.log(`💬 Twilio Conversations client created`);

    // 3. Wait for client to fully initialize
    await new Promise((resolve, reject) => {
      client.on("stateChanged", (state) => {
        console.log(`🌐 Client state changed to: ${state}`);
        if (state === "initialized") resolve();
        if (state === "failed") reject(new Error("Client initialization failed"));
      });

      client.on("connectionError", (error) => {
        console.error(`❌ Twilio client connection error: ${error.message}`);
      });
    });

    console.log(`✅ Twilio client fully initialized`);

    // 4. Create or join conversation
    // const convoRes = await fetch("/create-conversation");
    const convoRes = await fetch(`/create-conversation?identity=${identity}`);

    if (!convoRes.ok) throw new Error(`❌ Failed to create conversation`);
    const { conversationSid } = await convoRes.json();

    console.log(`💬 Conversation SID: ${conversationSid}`);
    conversation = await client.getConversationBySid(conversationSid);
    console.log(`✅ Joined conversation: ${conversation.sid}`);

    // 5. Listen for new messages
    conversation.on("messageAdded", (message) => {
      const messagesDiv = document.getElementById("messages");

      const author = message.author;
      const isMe = author === client.user.identity;
      const messageClass = isMe ? "my-message" : "their-message";

      const msgEl = document.createElement("p");
      msgEl.className = messageClass;
      msgEl.innerHTML = `<strong>${author}:</strong> ${message.body}`;
      messagesDiv.appendChild(msgEl);
      messagesDiv.scrollTop = messagesDiv.scrollHeight;

      // messagesDiv.innerHTML += `<p><strong>${message.author}:</strong> ${message.body}</p>`;
      messagesDiv.scrollTop = messagesDiv.scrollHeight;
    });
  } catch (error) {
    console.error(`🔥 Chat initialization error:`, error);
  }
}

async function sendMessage() {
  const input = document.getElementById("messageInput");
  const message = input.value.trim();
  if (!message) return;

  input.value = "";
  try {
    if (!conversation) throw new Error("❌ No conversation initialized");
    await conversation.sendMessage(message);
    console.log(`📤 Message sent: ${message}`);
  } catch (err) {
    console.error(`❌ Send message error: ${err.message}`);
  }
}
