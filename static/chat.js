let client;
let conversation;

async function initChat(identity) {
  // Fetch token
  const response = await fetch(`/token?identity=${identity}`);
  const { token } = await response.json();

  // Initialize Twilio Conversations client
  client = await Twilio.Conversations.Client.create(token);

  // Create or join conversation
  const convoResponse = await fetch("/create-conversation");
  const { conversationSid } = await convoResponse.json();

  conversation = await client.getConversationBySid(conversationSid);

  // Listen for new messages
  conversation.on("messageAdded", (message) => {
    const messagesDiv = document.getElementById("messages");
    messagesDiv.innerHTML += `<p><strong>${message.author}:</strong> ${message.body}</p>`;
  });
}

async function sendMessage() {
  const input = document.getElementById("messageInput");
  const message = input.value;
  input.value = "";

  if (conversation && message) {
    await conversation.sendMessage(message);
  }
}
