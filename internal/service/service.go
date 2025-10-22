package service

import (
	"context"
	"log"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/repository"
)

type Server struct {
	pb.UnimplementedOsmiServiceServer
	customerRepo *repository.CustomerRepository
	ticketRepo   *repository.TicketRepository
	eventRepo    *repository.EventRepository
}

func NewServer() *Server {
	return &Server{
		customerRepo: repository.NewCustomerRepository(),
		ticketRepo:   repository.NewTicketRepository(),
		eventRepo:    repository.NewEventRepository(),
	}
}

// CreateEvent con lógica real
func (s *Server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("CreateEvent called: %s (%s)", req.Name, req.Location)

	event := &repository.Event{
		PublicID:         "EVT-" + time.Now().Format("20060102150405"),
		Name:             req.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		StartDate:        parseDate(req.StartDate),
		EndDate:          parseDate(req.EndDate),
		Location:         req.Location,
		VenueDetails:     req.VenueDetails,
		Category:         req.Category,
		Tags:             req.Tags,
		IsActive:         req.IsActive,
		IsPublished:      req.IsPublished,
		ImageURL:         req.ImageUrl,
		BannerURL:        req.BannerUrl,
		MaxAttendees:     int32(req.MaxAttendees),
	}

	_, err := s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, err
	}

	return &pb.EventResponse{
		PublicId: event.PublicID,
		Name:     event.Name,
		Location: event.Location,
		Date:     req.StartDate,
	}, nil
}

// GetEvent con lógica real
func (s *Server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	log.Printf("GetEvent called: ID=%s", req.EventId)

	event, err := s.eventRepo.GetEventByPublicID(ctx, req.EventId)
	if err != nil {
		log.Printf("Error getting event: %v", err)
		return nil, err
	}

	return &pb.EventResponse{
		PublicId: event.PublicID,
		Name:     event.Name,
		Location: event.Location,
		Date:     event.StartDate.Format("2006-01-02"),
	}, nil
}

// ListEvents con lógica real
func (s *Server) ListEvents(ctx context.Context, _ *pb.Empty) (*pb.EventListResponse, error) {
	log.Println("ListEvents called")

	events, err := s.eventRepo.ListEvents(ctx)
	if err != nil {
		log.Printf("Error listing events: %v", err)
		return nil, err
	}

	var responses []*pb.EventResponse
	for _, e := range events {
		responses = append(responses, &pb.EventResponse{
			PublicId: e.PublicID,
			Name:     e.Name,
			Location: e.Location,
			Date:     e.StartDate.Format("2006-01-02"),
		})
	}

	return &pb.EventListResponse{Events: responses}, nil
}

func parseDate(input string) time.Time {
	t, err := time.Parse("2006-01-02", input)
	if err != nil {
		return time.Now()
	}
	return t
}
