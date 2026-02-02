package integration

import (
	"context"
	"testing"
	"time"

	// Ajusta a tus paquetes reales
	"yourmodule/internal/models"
	"yourmodule/internal/repository/testdb"
	"yourmodule/internal/services"
)

func newCtx(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

func TestFullFlow_EventCategoryCustomerTicket(t *testing.T) {
	db, cleanup := testdb.New(t)
	defer cleanup()

	// Instancia servicios con repos reales
	svc := services.New(db)

	ctx := newCtx(t, 10*time.Second)

	// 1) Evento
	ev, err := svc.Events.Create(ctx, models.Event{ /* campos requeridos */ })
	if err != nil {
		t.Fatalf("create event failed: %v", err)
	}

	// 2) Categoría
	cat, err := svc.Categories.Create(ctx, models.Category{ /* campos requeridos, vinculado a ev */ })
	if err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	// 3) Cliente
	cust, err := svc.Customers.Create(ctx, models.Customer{
		Name:  "Integration User",
		Email: "int-" + time.Now().Format("150405.000000000") + "@osmi-test.local",
	})
	if err != nil {
		t.Fatalf("create customer failed: %v", err)
	}

	// 4) Ticket
	tk, err := svc.Tickets.Create(ctx, models.TicketInput{
		EventID:    ev.PublicID,
		CategoryID: cat.PublicID,
		CustomerID: cust.PublicID,
		Quantity:   2,
	})
	if err != nil {
		t.Fatalf("create ticket failed: %v", err)
	}

	// 5) Consulta
	got, err := svc.Tickets.GetByID(ctx, tk.TicketID)
	if err != nil {
		t.Fatalf("get ticket failed: %v", err)
	}
	if got.TicketID != tk.TicketID {
		t.Fatalf("expected ticket %s, got %s", tk.TicketID, got.TicketID)
	}
}
