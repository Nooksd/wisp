// Arquivo: lib/main.dart
import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';
import 'dart:async';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:synchronized/synchronized.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  MessageManager.instance.initialize();
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Wisp Chat',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: const ContactsPage(),
    );
  }
}

// Modelo para usuário
class User {
  final String userId;
  final String name;
  final String email;

  User({required this.userId, required this.name, required this.email});
}

// Modelo para mensagem
class ChatMessage {
  final String content;
  final bool isSent;
  final DateTime timestamp;
  final String? msgId;

  ChatMessage({
    required this.content,
    required this.isSent,
    required this.timestamp,
    this.msgId,
  });
}

// Serviço de WebSocket
class WebSocketService {
  WebSocketChannel? _channel;
  final String _baseUrl = 'ws://192.168.1.20:9999/ws';
  final String _deviceId = '33123754';
  final String _token =
      'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJqb2FvMTIzIiwiaXNBZG1pbiI6ZmFsc2UsImV4cCI6MTc0NjQ2OTgzMywiaWF0IjoxNzQ1ODY1MDMzLCJqdGkiOiI2ODBmYzk0OTBjZTE4NjcyZjBiNTI0MWUifQ.VqnJS9XtMHE90IYTdo2BISGTUmezLvqmy53CKSrnEUw';
  final Function(Map<String, dynamic>) onMessageReceived;

  WebSocketService({required this.onMessageReceived});

  void connect() {
    try {
      _channel = WebSocketChannel.connect(
        Uri.parse('$_baseUrl?deviceId=$_deviceId&token=$_token'),
      );

      _channel!.stream.listen(
        (message) {
          final data = jsonDecode(message);
          onMessageReceived(data);
        },
        onDone: _reconnect,
        onError: (error) {
          _reconnect();
        },
      );
    } catch (e) {
      _reconnect();
    }
  }

  void _reconnect() {
    Timer(Duration(seconds: 3), connect);
  }

  void sendMessage(Map<String, dynamic> message) {
    if (_channel != null) {
      _channel!.sink.add(jsonEncode(message));
    }
  }

  void close() {
    _channel?.sink.close();
  }
}

// Gerenciador de mensagens
class MessageManager {
  final _lock = Lock();
  SharedPreferences? _prefs;
  final Map<String, List<ChatMessage>> _messages = {};
  late final WebSocketService _webSocketService;
  final User _currentUser;
  Function? onMessageReceived;
  static final MessageManager _instance = MessageManager._internal();

  MessageManager(this._currentUser)
      : _webSocketService = WebSocketService(onMessageReceived: (data) {
          if (data['type'] == 'message') {
            final message = ChatMessage(
              content: data['content'],
              isSent: false,
              timestamp: DateTime.now(),
              msgId: data['msgId'],
            );

            // Adicionar mensagem ao mapa
            final fromUserId = data['from'];
            if (!_instance._messages.containsKey(fromUserId)) {
              _instance._messages[fromUserId] = [];
            }
            _instance._messages[fromUserId]!.add(message);

            // Notificar UI
            if (_instance.onMessageReceived != null) {
              _instance.onMessageReceived!(fromUserId);
            }

            // Enviar recibo
            _instance._webSocketService.sendMessage({
              'type': 'receipt',
              'from': _instance._currentUser.userId,
              'to': fromUserId,
              'msgId': data['msgId'],
              'status': 'delivered',
            });
          }
        }) {
    _webSocketService.connect();
  }
  static MessageManager get instance => _instance;

  MessageManager._internal()
      : _currentUser = User(
          userId: 'joao123',
          name: 'Usuário Flutter',
          email: 'teste2@gmail.com',
        ) {
    _webSocketService =
        WebSocketService(onMessageReceived: _handleSocketMessage);
  }

  List<ChatMessage> getMessagesForUser(String userId) {
    return _messages[userId] ?? [];
  }

  Future<void> initialize() async {
    await _lock.synchronized(() async {
      _prefs = await SharedPreferences.getInstance();
      await _loadPersistedMessages();
      _webSocketService.connect();
    });
  }

  Future<void> _loadPersistedMessages() async {
    final messagesJson = _prefs?.getString('messages');
    if (messagesJson != null) {
      final Map<String, dynamic> data = jsonDecode(messagesJson);
      data.forEach((userId, messages) {
        _messages[userId] = (messages as List)
            .map((m) => ChatMessage(
                  content: m['content'],
                  isSent: m['isSent'],
                  timestamp: DateTime.parse(m['timestamp']),
                  msgId: m['msgId'],
                ))
            .toList();
      });
    }
  }

  void _persistMessages() async {
    final Map<String, dynamic> data = {};
    _messages.forEach((userId, messages) {
      data[userId] = messages
          .map((m) => {
                'content': m.content,
                'isSent': m.isSent,
                'timestamp': m.timestamp.toIso8601String(),
                'msgId': m.msgId,
              })
          .toList();
    });
    await _prefs?.setString('messages', jsonEncode(data));
  }

  void _handleSocketMessage(Map<String, dynamic> data) {
    if (data['type'] == 'message') {
      final message = ChatMessage(
        content: data['content'],
        isSent: false,
        timestamp:
            DateTime.fromMillisecondsSinceEpoch(data['timestamp'] * 1000),
        msgId: data['msgId'],
      );

      final fromUserId = data['from'];
      _addMessage(fromUserId, message);

      _webSocketService.sendMessage({
        'type': 'receipt',
        'from': _currentUser.userId,
        'to': fromUserId,
        'msgId': data['msgId'],
        'status': 'delivered',
      });
    }
  }

  void _addMessage(String userId, ChatMessage message) {
    _lock.synchronized(() {
      if (!_messages.containsKey(userId)) {
        _messages[userId] = [];
      }
      _messages[userId]!.add(message);
      _persistMessages();

      if (onMessageReceived != null) {
        onMessageReceived!(userId);
      }
    });
  }

  void sendMessage(String toUserId, String content) {
    final msgId =
        '${DateTime.now().millisecondsSinceEpoch}-${_currentUser.userId}';

    final message = ChatMessage(
      content: content,
      isSent: true,
      timestamp: DateTime.now(),
      msgId: msgId,
    );

    _addMessage(toUserId, message);

    _webSocketService.sendMessage({
      'type': 'message',
      'from': _currentUser.userId,
      'to': toUserId,
      'content': content,
      'msgId': msgId,
    });
  }

  void dispose() {
    _webSocketService.close();
  }
}

class ContactsPage extends StatelessWidget {
  const ContactsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final contacts = [
      User(
        userId: 'joao321',
        name: 'Usuário Web',
        email: 'teste@gmail.com',
      )
    ];

    return Scaffold(
      appBar: AppBar(
        title: Text('Contatos'),
      ),
      body: ListView.builder(
        itemCount: contacts.length,
        itemBuilder: (context, index) {
          final contact = contacts[index];
          return ListTile(
            title: Text(contact.name),
            subtitle: Text(contact.email),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => ChatPage(contact: contact),
                ),
              );
            },
          );
        },
      ),
    );
  }
}

class ChatPage extends StatefulWidget {
  final User contact;

  const ChatPage({super.key, required this.contact});

  @override
  ChatPageState createState() => ChatPageState();
}

class ChatPageState extends State<ChatPage> {
  final TextEditingController _controller = TextEditingController();
  late MessageManager _messageManager;
  List<ChatMessage> messages = [];

  @override
  void initState() {
    super.initState();
    _messageManager = MessageManager.instance;
    _messageManager.onMessageReceived = _onNewMessage;
    _loadMessages();
  }

  void _loadMessages() {
    setState(() {
      messages = _messageManager.getMessagesForUser(widget.contact.userId);
    });
  }

  void _onNewMessage(String fromUserId) {
    if (fromUserId == widget.contact.userId) {
      _loadMessages();
    }
  }

  void _sendMessage() {
    if (_controller.text.trim().isNotEmpty) {
      _messageManager.sendMessage(
        widget.contact.userId,
        _controller.text.trim(),
      );
      _controller.clear();
      _loadMessages();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.contact.name),
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              reverse: true,
              padding: EdgeInsets.all(8.0),
              itemCount: messages.length,
              itemBuilder: (context, index) {
                final reversedIndex = messages.length - 1 - index;
                final message = messages[reversedIndex];

                return Align(
                  alignment: message.isSent
                      ? Alignment.centerRight
                      : Alignment.centerLeft,
                  child: Container(
                    margin: EdgeInsets.symmetric(vertical: 4.0),
                    padding: EdgeInsets.all(8.0),
                    decoration: BoxDecoration(
                      color:
                          message.isSent ? Colors.blue[100] : Colors.grey[300],
                      borderRadius: BorderRadius.circular(12.0),
                    ),
                    child: Text(message.content),
                  ),
                );
              },
            ),
          ),
          Container(
            padding: EdgeInsets.symmetric(horizontal: 8.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _controller,
                    decoration: InputDecoration(
                      hintText: 'Digite uma mensagem...',
                    ),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.send),
                  onPressed: _sendMessage,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }
}
