package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
)

type SupportTicketRow struct {
	ID              string
	OrganizationID  string
	Number          int
	Title           string
	Category        string
	Priority        string
	Status          string
	SlaDueAt        *time.Time
	AssigneeUserID  *string
	RequesterUserID *string
	CustomerID      *string
	PhoneLineID     *string
	ChargeRef       *string
	InvoiceID       *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ResolvedAt      *time.Time
	ClosedAt        *time.Time
}

type SupportTicketMessageRow struct {
	ID                    string
	TicketID              string
	AuthorUserID          *string
	AuthorName            *string
	Visibility            string
	Body                  string
	AttachmentKey         *string
	AttachmentName        *string
	AttachmentBucket      *string
	AttachmentContentType *string
	AttachmentSizeBytes   *int64
	CreatedAt             time.Time
}

type SupportTicketHistoryRow struct {
	ID          string
	TicketID    string
	ActorUserID *string
	EventType   string
	FromValue   *string
	ToValue     *string
	Notes       *string
	CreatedAt   time.Time
}

func (s *Store) NextSupportTicketNumber(ctx context.Context, orgID string) (int, error) {
	var n int
	err := s.q(ctx).QueryRow(ctx, `SELECT COALESCE(MAX("Number"), 0) + 1 FROM "SupportTickets" WHERE "OrganizationId"=$1`, orgID).Scan(&n)
	return n, err
}

func (s *Store) InsertSupportTicket(ctx context.Context, r SupportTicketRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "SupportTickets" (
			"Id","OrganizationId","Number","Title","Category","Priority","Status","SlaDueAt",
			"AssigneeUserId","RequesterUserId","CustomerId","PhoneLineId","ChargeRef","InvoiceId",
			"CreatedAt","UpdatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		r.ID, r.OrganizationID, r.Number, r.Title, r.Category, r.Priority, r.Status, r.SlaDueAt,
		r.AssigneeUserID, r.RequesterUserID, r.CustomerID, r.PhoneLineID, r.ChargeRef, r.InvoiceID,
		r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *Store) UpdateSupportTicket(ctx context.Context, r SupportTicketRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "SupportTickets" SET
			"Title"=$2, "Category"=$3, "Priority"=$4, "Status"=$5, "SlaDueAt"=$6,
			"AssigneeUserId"=$7, "CustomerId"=$8, "PhoneLineId"=$9, "ChargeRef"=$10, "InvoiceId"=$11,
			"UpdatedAt"=$12, "ResolvedAt"=$13, "ClosedAt"=$14
		WHERE "Id"=$1 AND "OrganizationId"=$15`,
		r.ID, r.Title, r.Category, r.Priority, r.Status, r.SlaDueAt,
		r.AssigneeUserID, r.CustomerID, r.PhoneLineID, r.ChargeRef, r.InvoiceID,
		r.UpdatedAt, r.ResolvedAt, r.ClosedAt, r.OrganizationID)
	return err
}

func (s *Store) GetSupportTicket(ctx context.Context, orgID, id string) (*SupportTicketRow, error) {
	row := s.q(ctx).QueryRow(ctx, supportTicketSelect+` WHERE t."OrganizationId"=$1 AND t."Id"=$2`, orgID, id)
	item, err := scanSupportTicket(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListSupportTickets(ctx context.Context, orgID string, customerID *string, status string, page httputil.PageSearch) ([]SupportTicketRow, int64, error) {
	where := ` WHERE t."OrganizationId"=$1`
	args := []any{orgID}
	n := 2
	if customerID != nil && *customerID != "" {
		where += ` AND t."CustomerId"=$` + itoa(n)
		args = append(args, *customerID)
		n++
	}
	if status != "" {
		where += ` AND t."Status"=$` + itoa(n)
		args = append(args, status)
		n++
	}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM "SupportTickets" t`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, supportTicketSelect+where+` ORDER BY t."UpdatedAt" DESC OFFSET $`+itoa(n)+` LIMIT $`+itoa(n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []SupportTicketRow
	for rows.Next() {
		item, err := scanSupportTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []SupportTicketRow{}
	}
	return items, total, rows.Err()
}

func (s *Store) InsertSupportTicketMessage(ctx context.Context, m SupportTicketMessageRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "SupportTicketMessages" (
			"Id","TicketId","AuthorUserId","AuthorName","Visibility","Body",
			"AttachmentKey","AttachmentName","AttachmentBucket","AttachmentContentType","AttachmentSizeBytes","CreatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TicketID, m.AuthorUserID, m.AuthorName, m.Visibility, m.Body,
		m.AttachmentKey, m.AttachmentName, m.AttachmentBucket, m.AttachmentContentType, m.AttachmentSizeBytes, m.CreatedAt)
	return err
}

const supportTicketMessageSelect = `SELECT "Id","TicketId","AuthorUserId","AuthorName","Visibility","Body",
	"AttachmentKey","AttachmentName","AttachmentBucket","AttachmentContentType","AttachmentSizeBytes","CreatedAt"
	FROM "SupportTicketMessages"`

func scanSupportTicketMessage(row approvalScanner) (SupportTicketMessageRow, error) {
	var m SupportTicketMessageRow
	err := row.Scan(&m.ID, &m.TicketID, &m.AuthorUserID, &m.AuthorName, &m.Visibility, &m.Body,
		&m.AttachmentKey, &m.AttachmentName, &m.AttachmentBucket, &m.AttachmentContentType, &m.AttachmentSizeBytes, &m.CreatedAt)
	return m, err
}

func (s *Store) ListSupportTicketMessages(ctx context.Context, ticketID string, includeInternal bool) ([]SupportTicketMessageRow, error) {
	q := supportTicketMessageSelect + ` WHERE "TicketId"=$1`
	if !includeInternal {
		q += ` AND "Visibility"='public'`
	}
	q += ` ORDER BY "CreatedAt"`
	rows, err := s.q(ctx).Query(ctx, q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SupportTicketMessageRow
	for rows.Next() {
		item, err := scanSupportTicketMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []SupportTicketMessageRow{}
	}
	return items, rows.Err()
}

func (s *Store) GetSupportTicketMessage(ctx context.Context, ticketID, messageID string) (*SupportTicketMessageRow, error) {
	row := s.q(ctx).QueryRow(ctx, supportTicketMessageSelect+` WHERE "TicketId"=$1 AND "Id"=$2`, ticketID, messageID)
	item, err := scanSupportTicketMessage(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) InsertSupportTicketHistory(ctx context.Context, h SupportTicketHistoryRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "SupportTicketHistory" ("Id","TicketId","ActorUserId","EventType","FromValue","ToValue","Notes","CreatedAt")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		h.ID, h.TicketID, h.ActorUserID, h.EventType, h.FromValue, h.ToValue, h.Notes, h.CreatedAt)
	return err
}

func (s *Store) ListSupportTicketHistory(ctx context.Context, ticketID string) ([]SupportTicketHistoryRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id","TicketId","ActorUserId","EventType","FromValue","ToValue","Notes","CreatedAt"
		FROM "SupportTicketHistory" WHERE "TicketId"=$1 ORDER BY "CreatedAt"`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SupportTicketHistoryRow
	for rows.Next() {
		var h SupportTicketHistoryRow
		if err := rows.Scan(&h.ID, &h.TicketID, &h.ActorUserID, &h.EventType, &h.FromValue, &h.ToValue, &h.Notes, &h.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	if items == nil {
		items = []SupportTicketHistoryRow{}
	}
	return items, rows.Err()
}

const supportTicketSelect = `
	SELECT t."Id", t."OrganizationId", t."Number", t."Title", t."Category", t."Priority", t."Status",
		t."SlaDueAt", t."AssigneeUserId", t."RequesterUserId", t."CustomerId", t."PhoneLineId",
		t."ChargeRef", t."InvoiceId", t."CreatedAt", t."UpdatedAt", t."ResolvedAt", t."ClosedAt"
	FROM "SupportTickets" t`

func scanSupportTicket(row approvalScanner) (SupportTicketRow, error) {
	var r SupportTicketRow
	err := row.Scan(&r.ID, &r.OrganizationID, &r.Number, &r.Title, &r.Category, &r.Priority, &r.Status,
		&r.SlaDueAt, &r.AssigneeUserID, &r.RequesterUserID, &r.CustomerID, &r.PhoneLineID,
		&r.ChargeRef, &r.InvoiceID, &r.CreatedAt, &r.UpdatedAt, &r.ResolvedAt, &r.ClosedAt)
	return r, err
}
