package store

import "context"

func (s *Store) InsertExternalLineImportJob(ctx context.Context, id, orgID, source string, providerID *string, bucket, key string, fileName *string, layout string, refPeriod *string, partial bool, errMsg string, actor *string, createdAt interface{}) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ExternalLineImportJobs" (
			"Id","OrganizationId","Source","ProviderId","Status","StorageBucket","StorageObjectKey",
			"FileName","LayoutCode","ReferencePeriod","IsPartialSource","ErrorMessage","CreatedByUserId","CreatedAt"
		) VALUES ($1,$2,$3,$4,'awaiting_layout',$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, orgID, source, providerID, bucket, key, fileName, layout, refPeriod, partial, errMsg, actor, createdAt)
	return err
}
