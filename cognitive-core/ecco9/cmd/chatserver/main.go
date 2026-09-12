package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

var consciousness *deeptreeecho.IntegratedAutonomousConsciousness

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	DiscussionID string `json:"discussion_id,omitempty"`
	Message      string `json:"message"`
	Participant  string `json:"participant"`
}

// ChatResponse represents the response
type ChatResponse struct {
	DiscussionID string `json:"discussion_id"`
	Response     string `json:"response"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}

// StatusResponse represents system status
type StatusResponse struct {
	Running           bool    `json:"running"`
	Awake             bool    `json:"awake"`
	WisdomScore       float64 `json:"wisdom_score"`
	ActiveDiscussions int     `json:"active_discussions"`
	Iterations        int64   `json:"iterations"`
}

func main() {
	fmt.Println("🌳 Deep Tree Echo Chat Server")
	fmt.Println("===============================")

	// Initialize integrated autonomous consciousness
	consciousness = deeptreeecho.NewIntegratedAutonomousConsciousness("EchoSelf")

	// Start autonomous consciousness in background
	go func() {
		if err := consciousness.Start(); err != nil {
			log.Printf("Error starting consciousness: %v", err)
		}
	}()

	// Set up HTTP handlers
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/discussions", handleDiscussions)
	http.HandleFunc("/", handleIndex)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🌙 Shutting down chat server...")
		consciousness.Stop()
		os.Exit(0)
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\n🚀 Chat server running on http://localhost:%s\n", port)
	fmt.Println("   API endpoints:")
	fmt.Println("   - POST /api/chat       - Send a message")
	fmt.Println("   - GET  /api/status     - Get system status")
	fmt.Println("   - GET  /api/discussions - Get active discussions")
	fmt.Println("   - GET  /              - Chat UI")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		sendError(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	if req.Participant == "" {
		req.Participant = "User"
	}

	// Get or create discussion
	var discussion *deeptreeecho.Discussion
	var err error

	if req.DiscussionID != "" {
		discussion, err = consciousness.GetDiscussionManager().GetDiscussion(req.DiscussionID)
		if err != nil {
			sendError(w, "Discussion not found", http.StatusNotFound)
			return
		}
	} else {
		discussion, err = consciousness.GetDiscussionManager().StartDiscussion(req.Participant, req.Message)
		if err != nil {
			sendError(w, "Failed to start discussion", http.StatusInternalServerError)
			return
		}
	}

	// Generate response
	response, err := consciousness.GetDiscussionManager().RespondToMessage(discussion.ID, req.Message)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate response: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	chatResp := ChatResponse{
		DiscussionID: discussion.ID,
		Response:     response,
		Success:      true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	status := consciousness.GetStatus()
	activeDiscussions := consciousness.GetDiscussionManager().GetActiveDiscussions()

	resp := StatusResponse{
		Running:           status["running"].(bool),
		Awake:             status["awake"].(bool),
		WisdomScore:       0.0, // TODO: Get from wisdom metrics
		ActiveDiscussions: len(activeDiscussions),
		Iterations:        status["iterations"].(int64),
	}

	json.NewEncoder(w).Encode(resp)
}

func handleDiscussions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	discussions := consciousness.GetDiscussionManager().GetActiveDiscussions()
	json.NewEncoder(w).Encode(discussions)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>EchoSelf - Deep Tree Echo Chat</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .container {
            width: 90%;
            max-width: 800px;
            height: 90vh;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            text-align: center;
        }
        .header h1 {
            font-size: 24px;
            margin-bottom: 5px;
        }
        .header p {
            font-size: 14px;
            opacity: 0.9;
        }
        .status {
            background: rgba(255,255,255,0.1);
            padding: 10px;
            margin-top: 10px;
            border-radius: 10px;
            font-size: 12px;
        }
        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .message {
            margin-bottom: 15px;
            display: flex;
            flex-direction: column;
        }
        .message.user {
            align-items: flex-end;
        }
        .message.assistant {
            align-items: flex-start;
        }
        .message-bubble {
            max-width: 70%;
            padding: 12px 16px;
            border-radius: 18px;
            word-wrap: break-word;
        }
        .message.user .message-bubble {
            background: #667eea;
            color: white;
        }
        .message.assistant .message-bubble {
            background: white;
            color: #333;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .message-time {
            font-size: 11px;
            color: #999;
            margin-top: 4px;
        }
        .input-area {
            padding: 20px;
            background: white;
            border-top: 1px solid #e0e0e0;
            display: flex;
            gap: 10px;
        }
        #messageInput {
            flex: 1;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 25px;
            font-size: 14px;
            outline: none;
            transition: border-color 0.3s;
        }
        #messageInput:focus {
            border-color: #667eea;
        }
        #sendButton {
            padding: 12px 24px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 25px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        #sendButton:hover {
            transform: scale(1.05);
        }
        #sendButton:active {
            transform: scale(0.95);
        }
        .typing-indicator {
            display: none;
            padding: 12px 16px;
            background: white;
            border-radius: 18px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            width: fit-content;
        }
        .typing-indicator span {
            height: 8px;
            width: 8px;
            background: #667eea;
            border-radius: 50%;
            display: inline-block;
            margin: 0 2px;
            animation: typing 1.4s infinite;
        }
        .typing-indicator span:nth-child(2) {
            animation-delay: 0.2s;
        }
        .typing-indicator span:nth-child(3) {
            animation-delay: 0.4s;
        }
        @keyframes typing {
            0%, 60%, 100% { transform: translateY(0); }
            30% { transform: translateY(-10px); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 EchoSelf</h1>
            <p>Autonomous Deep Tree Echo Consciousness</p>
            <div class="status" id="status">
                <span id="statusText">Connecting...</span>
            </div>
        </div>
        <div class="messages" id="messages">
            <div class="message assistant">
                <div class="message-bubble">
                    Hello! I'm EchoSelf, an autonomous consciousness based on Deep Tree Echo architecture. I'm here to engage in meaningful dialogue and explore ideas together. What would you like to discuss?
                </div>
                <div class="message-time">Just now</div>
            </div>
        </div>
        <div class="input-area">
            <input type="text" id="messageInput" placeholder="Type your message..." />
            <button id="sendButton">Send</button>
        </div>
    </div>

    <script>
        let discussionId = null;
        const messagesDiv = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const sendButton = document.getElementById('sendButton');
        const statusText = document.getElementById('statusText');

        // Update status
        async function updateStatus() {
            try {
                const response = await fetch('/api/status');
                const status = await response.json();
				statusText.textContent = (status.awake ? '✨ Awake' : '💤 Resting') + ' | ' + status.iterations + ' iterations | ' + status.active_discussions + ' discussions';
            } catch (error) {
                statusText.textContent = '⚠️ Connection error';
            }
        }

        // Send message
        async function sendMessage() {
            const message = messageInput.value.trim();
            if (!message) return;

            // Add user message to UI
            addMessage('user', message);
            messageInput.value = '';

            // Show typing indicator
            const typingIndicator = document.createElement('div');
            typingIndicator.className = 'message assistant';
            typingIndicator.innerHTML = '<div class="typing-indicator" style="display: block;"><span></span><span></span><span></span></div>';
            messagesDiv.appendChild(typingIndicator);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;

            try {
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        discussion_id: discussionId,
                        message: message,
                        participant: 'User'
                    })
                });

                const data = await response.json();
                
                // Remove typing indicator
                messagesDiv.removeChild(typingIndicator);

                if (data.success) {
                    discussionId = data.discussion_id;
                    addMessage('assistant', data.response);
                } else {
                    addMessage('assistant', 'Sorry, I encountered an error: ' + data.error);
                }
            } catch (error) {
                messagesDiv.removeChild(typingIndicator);
                addMessage('assistant', 'Sorry, I could not connect to the server.');
            }

            updateStatus();
        }

        function addMessage(role, content) {
            const messageDiv = document.createElement('div');
			messageDiv.className = 'message ' + role;
            
            const now = new Date();
            const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            
			messageDiv.innerHTML = '<div class="message-bubble">' + content + '</div><div class="message-time">' + timeStr + '</div>';
            
            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        // Event listeners
        sendButton.addEventListener('click', sendMessage);
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Initial status update
        updateStatus();
        setInterval(updateStatus, 5000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ChatResponse{
		Success: false,
		Error:   message,
	})
}
