export class WebSocketService {
    constructor(url) {
        this.url = url;
        this.socket = null;
        this.listeners = [];
    }

    connect() {
        this.socket = new WebSocket(this.url);
        this.socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.listeners.forEach(cb => cb(data));
            } catch (e) {
                console.error("WS Parse Error", e);
            }
        };
        this.socket.onopen = () => console.log("WS Connected");
        this.socket.onclose = () => console.log("WS Closed");
        this.socket.onerror = (e) => console.error("WS Error", e);
    }

    onMessage(callback) {
        this.listeners.push(callback);
    }

    close() {
        if (this.socket) this.socket.close();
    }
}
