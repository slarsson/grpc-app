package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"database/sql"

	"github.com/slarsson/grpc-app/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/mattn/go-sqlite3"
)

const PORT = ":5000"

type server struct {
	user.UnimplementedUserServiceServer
	db *sql.DB
}

func (s *server) Get(ctx context.Context, input *user.Id) (*user.User, error) {
	if input == nil {
		return nil, fmt.Errorf("noop")
	}

	var id int
	var email string
	var createdAt time.Time
	var updatedAt time.Time

	err := s.db.QueryRowContext(ctx, "SELECT user_id, email, created_at, updated_at FROM users WHERE user_id = ?", input.Id).
		Scan(&id, &email, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan: %v", err)
	}

	return &user.User{
		Id:        fmt.Sprintf("%d", id),
		Email:     ptr(email),
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}

func (s *server) Create(ctx context.Context, input *user.User) (*user.User, error) {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (email, updated_at, created_at) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Printf("prepare: %v", err)
		return nil, fmt.Errorf("prepare: %v", err)
	}

	now := time.Now()

	res, err := stmt.ExecContext(ctx, *input.Email, now, now)
	if err != nil {
		fmt.Printf("exec: %v", err)
		return nil, fmt.Errorf("exec: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("lastInsertedId: %v", err)
	}

	return &user.User{
		Id:        fmt.Sprintf("%d", id),
		Email:     input.Email,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}, nil
}

func main() {
	db, err := sql.Open("sqlite3", "./data.sqlite")
	if err != nil {
		log.Fatalf("sqlite3: %v", err)
	}

	fmt.Println("data.sqlite loaded")

	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()

	user.RegisterUserServiceServer(s, &server{db: db})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func ptr[T any](v T) *T {
	return &v
}
