package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

var ticketStatuses = map[string]struct{}{
	"aberto": {}, "em_triagem": {}, "em_atendimento": {}, "aguardando_cliente": {},
	"aguardando_operadora": {}, "resolvido": {}, "encerrado": {}, "reaberto": {},
}

func (s *Service) CreateSupportTicket(ctx context.Context, input models.CreateSupportTicketInput) (*models.SupportTicketResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, httputil.ValidationError(notifications.N("TICKET_TITLE_REQUIRED", "Informe o título do ticket."))
	}
	now := time.Now().UTC()
	n, err := s.Store.NextSupportTicketNumber(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	priority := firstNonEmpty(input.Priority, "media")
	sla := slaForPriority(priority, now)
	row := store.SupportTicketRow{
		ID: uuid.New().String(), OrganizationID: orgID, Number: n, Title: title,
		Category: firstNonEmpty(input.Category, "geral"), Priority: priority, Status: "aberto",
		SlaDueAt: &sla, RequesterUserID: &user.ID, CustomerID: optStr(input.CustomerID),
		PhoneLineID: optStr(input.PhoneLineID), ChargeRef: optStr(input.ChargeRef), InvoiceID: optStr(input.InvoiceID),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.Store.InsertSupportTicket(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.Store.InsertSupportTicketHistory(ctx, store.SupportTicketHistoryRow{
		ID: uuid.New().String(), TicketID: row.ID, ActorUserID: &user.ID, EventType: "created", ToValue: strPtr("aberto"), CreatedAt: now,
	})
	if body := strings.TrimSpace(input.Message); body != "" {
		_ = s.Store.InsertSupportTicketMessage(ctx, store.SupportTicketMessageRow{
			ID: uuid.New().String(), TicketID: row.ID, AuthorUserID: &user.ID, AuthorName: &user.Name,
			Visibility: "public", Body: body, CreatedAt: now,
		})
	}
	return s.GetSupportTicket(ctx, row.ID, true)
}

func (s *Service) ListSupportTickets(ctx context.Context, customerID, status string, page httputil.PageSearch) ([]models.SupportTicketResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	var cid *string
	if customerID != "" {
		cid = &customerID
	}
	rows, total, err := s.Store.ListSupportTickets(ctx, orgID, cid, status, page)
	if err != nil {
		return nil, 0, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	out := make([]models.SupportTicketResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, ticketToDTO(r, nil, nil))
	}
	return out, total, nil
}

func (s *Service) GetSupportTicket(ctx context.Context, id string, includeInternal bool) (*models.SupportTicketResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetSupportTicket(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		return nil, httputil.NotFoundError(notifications.N("TICKET_NOT_FOUND", "Ticket não encontrado."))
	}
	msgs, _ := s.Store.ListSupportTicketMessages(ctx, id, includeInternal)
	hist, _ := s.Store.ListSupportTicketHistory(ctx, id)
	dto := ticketToDTO(*row, msgs, hist)
	return &dto, nil
}

func (s *Service) UpdateSupportTicket(ctx context.Context, id string, input models.UpdateSupportTicketInput) (*models.SupportTicketResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetSupportTicket(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("TICKET_NOT_FOUND", "Ticket não encontrado."))
	}
	now := time.Now().UTC()
	if input.Title != nil {
		row.Title = strings.TrimSpace(*input.Title)
	}
	if input.Category != nil {
		row.Category = *input.Category
	}
	if input.Priority != nil {
		row.Priority = *input.Priority
		sla := slaForPriority(row.Priority, now)
		row.SlaDueAt = &sla
	}
	if input.AssigneeUserID != nil {
		row.AssigneeUserID = optStr(*input.AssigneeUserID)
		_ = s.Store.InsertSupportTicketHistory(ctx, store.SupportTicketHistoryRow{
			ID: uuid.New().String(), TicketID: row.ID, ActorUserID: &user.ID, EventType: "assign", ToValue: input.AssigneeUserID, CreatedAt: now,
		})
	}
	if input.Status != nil {
		st := strings.TrimSpace(*input.Status)
		if _, ok := ticketStatuses[st]; !ok {
			return nil, httputil.ValidationError(notifications.N("TICKET_STATUS_INVALID", "Status de ticket inválido."))
		}
		from := row.Status
		row.Status = st
		if st == "resolvido" {
			row.ResolvedAt = &now
		}
		if st == "encerrado" {
			row.ClosedAt = &now
		}
		if st == "reaberto" {
			row.ResolvedAt = nil
			row.ClosedAt = nil
		}
		_ = s.Store.InsertSupportTicketHistory(ctx, store.SupportTicketHistoryRow{
			ID: uuid.New().String(), TicketID: row.ID, ActorUserID: &user.ID, EventType: "status", FromValue: &from, ToValue: &st, CreatedAt: now,
		})
	}
	row.UpdatedAt = now
	if err := s.Store.UpdateSupportTicket(ctx, *row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetSupportTicket(ctx, id, true)
}

func (s *Service) AddSupportTicketMessage(ctx context.Context, id string, input models.AddSupportTicketMessageInput) (*models.SupportTicketResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetSupportTicket(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("TICKET_NOT_FOUND", "Ticket não encontrado."))
	}
	body := strings.TrimSpace(input.Body)
	key := strings.TrimSpace(input.AttachmentKey)
	name := strings.TrimSpace(input.AttachmentName)
	if body == "" && key == "" {
		return nil, httputil.ValidationError(notifications.TicketMessageEmpty)
	}
	if key != "" {
		if !validTicketObjectKey(orgID, id, key) {
			return nil, httputil.ValidationError(notifications.ObjectKeyInvalid)
		}
		if name == "" {
			return nil, httputil.ValidationError(notifications.TicketFilenameInvalid)
		}
	}
	vis := firstNonEmpty(input.Visibility, "public")
	now := time.Now().UTC()
	if err := s.Store.InsertSupportTicketMessage(ctx, store.SupportTicketMessageRow{
		ID: uuid.New().String(), TicketID: id, AuthorUserID: &user.ID, AuthorName: &user.Name,
		Visibility: vis, Body: body, AttachmentKey: optStr(key), AttachmentName: optStr(name),
		AttachmentBucket:      optStr(strings.TrimSpace(input.AttachmentBucket)),
		AttachmentContentType: optStr(strings.TrimSpace(input.AttachmentContentType)),
		AttachmentSizeBytes:   input.AttachmentSizeBytes, CreatedAt: now,
	}); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	row.UpdatedAt = now
	_ = s.Store.UpdateSupportTicket(ctx, *row)
	return s.GetSupportTicket(ctx, id, true)
}

func (s *Service) PrepareTicketAttachmentKey(ctx context.Context, ticketID string, input models.TicketAttachmentUploadInput, portal bool) (objectKey string, err error) {
	if _, err := s.requireTicketAccess(ctx, ticketID, portal); err != nil {
		return "", err
	}
	name := sanitizeTicketFileName(input.FileName)
	if name == "" {
		return "", httputil.ValidationError(notifications.TicketFilenameInvalid)
	}
	orgID, err := orgFrom(ctx)
	if err != nil {
		return "", err
	}
	return "tickets/" + orgID + "/" + ticketID + "/" + uuid.New().String() + "_" + name, nil
}

func (s *Service) TicketAttachmentLocation(ctx context.Context, ticketID, messageID string, portal bool) (*store.SupportTicketMessageRow, error) {
	if _, err := s.requireTicketAccess(ctx, ticketID, portal); err != nil {
		return nil, err
	}
	msg, err := s.Store.GetSupportTicketMessage(ctx, ticketID, messageID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if msg == nil || msg.AttachmentKey == nil || strings.TrimSpace(*msg.AttachmentKey) == "" {
		return nil, httputil.NotFoundError(notifications.TicketAttachmentNotFound)
	}
	if portal && msg.Visibility != "public" {
		return nil, httputil.NotFoundError(notifications.TicketAttachmentNotFound)
	}
	return msg, nil
}

func (s *Service) requireTicketAccess(ctx context.Context, ticketID string, portal bool) (*store.SupportTicketRow, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetSupportTicket(ctx, orgID, ticketID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		return nil, httputil.NotFoundError(notifications.N("TICKET_NOT_FOUND", "Ticket não encontrado."))
	}
	if portal {
		me, err := s.PortalGetMe(ctx)
		if err != nil {
			return nil, err
		}
		if row.CustomerID == nil || *row.CustomerID != me.Customer.ID {
			return nil, httputil.NotFoundError(notifications.N("TICKET_NOT_FOUND", "Ticket não encontrado."))
		}
	}
	return row, nil
}

func (s *Service) PortalGetSupportTicket(ctx context.Context, id string) (*models.SupportTicketResponse, error) {
	if _, err := s.requireTicketAccess(ctx, id, true); err != nil {
		return nil, err
	}
	return s.GetSupportTicket(ctx, id, false)
}

func (s *Service) PortalAddSupportTicketMessage(ctx context.Context, id string, input models.AddSupportTicketMessageInput) (*models.SupportTicketResponse, error) {
	if _, err := s.requireTicketAccess(ctx, id, true); err != nil {
		return nil, err
	}
	input.Visibility = "public"
	return s.AddSupportTicketMessage(ctx, id, input)
}

func sanitizeTicketFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "._")
	if out == "" || out == "." || out == ".." {
		return ""
	}
	if len(out) > 180 {
		out = out[:180]
	}
	return out
}

func validTicketObjectKey(orgID, ticketID, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "\x00") {
		return false
	}
	prefix := "tickets/" + orgID + "/" + ticketID + "/"
	return strings.HasPrefix(key, prefix)
}

func slaForPriority(priority string, now time.Time) time.Time {
	hours := 24
	switch strings.ToLower(priority) {
	case "critica", "alta":
		hours = 8
	case "baixa":
		hours = 72
	}
	return now.Add(time.Duration(hours) * time.Hour)
}

func ticketToDTO(r store.SupportTicketRow, msgs []store.SupportTicketMessageRow, hist []store.SupportTicketHistoryRow) models.SupportTicketResponse {
	out := models.SupportTicketResponse{
		ID: r.ID, Number: r.Number, Title: r.Title, Category: r.Category, Priority: r.Priority,
		Status: r.Status, SlaDueAt: r.SlaDueAt, AssigneeUserID: r.AssigneeUserID, RequesterUserID: r.RequesterUserID,
		CustomerID: r.CustomerID, PhoneLineID: r.PhoneLineID, ChargeRef: r.ChargeRef, InvoiceID: r.InvoiceID,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, ResolvedAt: r.ResolvedAt, ClosedAt: r.ClosedAt,
	}
	for _, m := range msgs {
		out.Messages = append(out.Messages, models.SupportTicketMessageResponse{
			ID: m.ID, AuthorUserID: m.AuthorUserID, AuthorName: m.AuthorName, Visibility: m.Visibility,
			Body: m.Body, AttachmentKey: m.AttachmentKey, AttachmentName: m.AttachmentName,
			AttachmentBucket: m.AttachmentBucket, AttachmentContentType: m.AttachmentContentType,
			AttachmentSizeBytes: m.AttachmentSizeBytes, CreatedAt: m.CreatedAt,
		})
	}
	for _, h := range hist {
		out.History = append(out.History, models.SupportTicketHistoryResponse{
			ID: h.ID, ActorUserID: h.ActorUserID, EventType: h.EventType, FromValue: h.FromValue, ToValue: h.ToValue, Notes: h.Notes, CreatedAt: h.CreatedAt,
		})
	}
	if out.Messages == nil {
		out.Messages = []models.SupportTicketMessageResponse{}
	}
	if out.History == nil {
		out.History = []models.SupportTicketHistoryResponse{}
	}
	return out
}
