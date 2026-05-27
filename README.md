# XPLORE

A modern, real-time chat application built with Go (Fiber) backend and React (Vite) frontend, featuring WebSocket-based real-time communication, Supabase authentication, and a beautiful wooden-themed UI.

## Features

- 🌲 **Beautiful Interface**: Unique wooden-themed UI with glassmorphism effects
- 💬 **Real-time Messaging**: Instant message delivery via WebSocket connections
- 🔒 **Secure Authentication**: Supabase-based authentication with JWT tokens
- 👥 **Group Chats & Direct Messages**: Create public/private rooms and direct message friends
- 📁 **File Sharing**: Upload and share images and files within chats
- ✅ **Read Receipts**: See when your messages have been read
- ⌨️ **Typing Indicators**: See when others are typing in real-time
- 🟢 **Presence Indicators**: See who's online and offline
- 🔍 **User Search & Friend System**: Find users and send friend requests
- 🚪 **Room Management**: Browse public rooms, request to join, and manage memberships
- 📱 **Responsive Design**: Works on desktop and mobile devices

## Tech Stack

### Backend
- **Go 1.22+** - Core language
- **Fiber v2** - High-performance web framework
- **PostgreSQL** - Primary database (via Supabase)
- **WebSocket** - Real-time communication
- **Supabase** - Authentication and file storage

### Frontend
- **React 18+** - UI library
- **Vite** - Build tool and development server
- **Zustand** - State management
- **Axios** - HTTP client
- **Supabase JS** - Authentication client

## Prerequisites

- Go 1.22 or higher
- Node.js 18 or higher
- Supabase account (for authentication and storage)
- PostgreSQL database (Supabase provides this)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/dinalegw/XPLORE.git
cd XPLORE
```

### 2. Set Up Environment Variables

Create a `.env` file in the `backend` directory:

```env
# Supabase Configuration
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_SERVICE_KEY=your_supabase_service_key
SUPABASE_DB_URL=your_supabase_database_url

# Server Configuration
PORT=8080
```

Create a `.env` file in the `frontend` directory:

```env
VITE_SUPABASE_URL=your_supabase_url
VITE_SUPABASE_ANON_KEY=your_supabase_anon_key
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

### 3. Install Dependencies

#### Backend
```bash
cd backend
go mod download
```

#### Frontend
```bash
cd ../frontend
npm install
```

### 4. Database Setup

Run the SQL migrations found in `database/schema.sql` to set up the required tables.

### 5. Run the Application

#### Backend
```bash
# From backend directory
go run main.go
```

#### Frontend
```bash
# From frontend directory
npm run dev
```

The application will be available at:
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080

## API Endpoints

### Authentication
- `GET /api/rooms` - Get user's rooms
- `GET /api/rooms/browse` - Browse public rooms
- `POST /api/rooms` - Create a new room
- `GET /api/rooms/requests` - Get join requests for rooms you own
- `POST /api/rooms/:id/request-join` - Request to join a room
- `POST /api/rooms/requests/:id/respond` - Respond to a join request
- `POST /api/rooms/:id/join` - Join a room (after approval)
- `GET /api/rooms/:id/messages` - Get messages for a room
- `POST /api/upload` - Upload a file
- `GET /api/users/search` - Search for users
- `POST /api/users/friend-request` - Send a friend request
- `GET /api/users/friend-requests` - Get friend requests
- `POST /api/users/friend-requests/:id/respond` - Respond to a friend request

### WebSocket
- `GET /ws` - WebSocket endpoint (requires `user_id` and `room_id` query parameters)

#### WebSocket Events
- `message.send` - Send a message to a room
- `message.new` - New message received in a room
- `typing.start` - User started typing
- `typing.stop` - User stopped typing
- `receipt.read` - Message read receipt
- `presence.online` - User came online
- `presence.offline` - User went offline

## Database Schema

The application requires the following tables:

```sql
-- Profiles table
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id),
    username TEXT NOT NULL,
    avatar_url TEXT,
    is_online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMP WITH TIME ZONE
);

-- Rooms table
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES profiles(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Room members table
CREATE TABLE room_members (
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

-- Join requests table
CREATE TABLE join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, user_id)
);

-- Messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    content TEXT,
    file_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Read receipts table
CREATE TABLE read_receipts (
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

-- Friend requests table
CREATE TABLE friend_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    receiver_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(sender_id, receiver_id)
);
```

## Environment Variables

### Backend (.env)
| Variable | Description |
|----------|-------------|
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_ANON_KEY` | Supabase anonymous key |
| `SUPABASE_SERVICE_KEY` | Supabase service key (keep secret) |
| `SUPABASE_DB_URL` | Supabase PostgreSQL connection string |
| `PORT` | Server port (default: 8080) |

### Frontend (.env)
| Variable | Description |
|----------|-------------|
| `VITE_SUPABASE_URL` | Supabase project URL |
| `VITE_SUPABASE_ANON_KEY` | Supabase anonymous key |
| `VITE_API_URL` | Backend API URL |
| `VITE_WS_URL` | WebSocket server URL |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Supabase](https://supabase.com) for authentication and storage
- [Go Fiber](https://gofiber.io) for the high-performance web framework
- [React](https://reactjs.org) and [Vite](https://vitejs.dev) for the frontend stack
- [Zustand](https://zustand-demo.pmndrs.net) for state management

## Contact

Daniel O Inalegwu - [dinalegw](https://github.com/dinalegw)

Project Link: [https://github.com/dinalegw/XPLORE](https://github.com/dinalegw/XPLORE)