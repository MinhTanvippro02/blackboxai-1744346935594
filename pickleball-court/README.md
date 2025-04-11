# PickleCourt - Pickleball Court Management System

A web-based application for managing pickleball courts, bookings, and training sessions. Built with Go, SQLite, and Tailwind CSS.

## Features

- **User Management**
  - Multiple user roles (Admin, Coach, Player)
  - User registration and authentication
  - Profile management

- **Court Management**
  - Add, edit, and delete courts
  - View court availability
  - Real-time booking system

- **Booking System**
  - Court reservations
  - Training session scheduling
  - Booking history tracking

- **Training Sessions**
  - Coach-led training sessions
  - Student enrollment
  - Session management

## Tech Stack

- **Backend**: Go (Gin Framework)
- **Database**: SQLite
- **Frontend**: HTML, Tailwind CSS, JavaScript
- **Authentication**: Session-based with secure cookie storage

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/pickleball-court.git
cd pickleball-court
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

The application will be available at `http://localhost:8000`

## Project Structure

```
pickleball-court/
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   │   ├── admin.go
│   │   ├── auth.go
│   │   ├── coach.go
│   │   └── player.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   ├── booking.go
│   │   ├── court.go
│   │   ├── database.go
│   │   └── user.go
│   └── routes/
│       └── routes.go
├── static/
│   ├── css/
│   ├── js/
│   └── images/
├── templates/
│   ├── admin_dashboard.html
│   ├── coach_dashboard.html
│   ├── error.html
│   ├── home.html
│   ├── layout.html
│   ├── login.html
│   ├── player_dashboard.html
│   ├── profile.html
│   └── register.html
├── go.mod
├── go.sum
└── README.md
```

## User Roles

### Admin
- Manage courts
- Manage users
- View all bookings
- System configuration

### Coach
- Create training sessions
- Manage training schedule
- View enrolled students

### Player
- Book courts
- Enroll in training sessions
- View booking history
- Manage profile

## Environment Variables

The application uses the following environment variables:

- `PORT`: Server port (default: 8000)
- `SESSION_SECRET`: Secret key for session encryption
- `ENV`: Environment mode (development/production)

## Development

To run the application in development mode:

```bash
go run main.go
```

For hot reloading during development, you can use tools like `air`:

```bash
air
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Courts Table
```sql
CREATE TABLE courts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Bookings Table
```sql
CREATE TABLE bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    court_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status TEXT NOT NULL,
    booking_type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (court_id) REFERENCES courts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Training Sessions Table
```sql
CREATE TABLE training_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    coach_id INTEGER NOT NULL,
    court_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    max_participants INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (coach_id) REFERENCES users(id),
    FOREIGN KEY (court_id) REFERENCES courts(id)
);
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Tailwind CSS](https://tailwindcss.com)
- [Font Awesome](https://fontawesome.com)
