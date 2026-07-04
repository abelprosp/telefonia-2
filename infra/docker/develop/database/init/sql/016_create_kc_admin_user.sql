BEGIN;

WITH new_user AS (
  INSERT INTO user_entity (
    id, username, email, email_constraint, email_verified,
    enabled, realm_id, created_timestamp, not_before
  )
  SELECT
    gen_random_uuid()::text,
    'sysadmin',
    'sysadmin@luxus.com.br',
    lower('sysadmin@luxus.com.br'),
    true,
    true,
    r.id,
    (extract(epoch from now()) * 1000)::bigint,
    0
  FROM realm r WHERE r.name = 'master'
  RETURNING id
)
INSERT INTO credential (
  id, type, user_id, created_date, user_label, priority,
  secret_data, credential_data
)
SELECT
  gen_random_uuid()::text,
  'password',
  new_user.id,
  (extract(epoch from now()) * 1000)::bigint,
  'Initial password',
  10,
  '{"value":"zySt8sy4h7LyrNPapGVtBo40zaoqd6JL/q5sgFoCEw5q7VA1hmlxl0t6orFoa6eUui/XkCJhqI5wEYbDw3Ta0A==","salt":"2Fu5Vuylp/1ix2YeRNK94w==","additionalParameters":{}}',
  '{"hashIterations":27500,"algorithm":"pbkdf2-sha256","additionalParameters":{}}'
FROM new_user;

-- Grant full admin rights (the master realm's "admin" role)
INSERT INTO user_role_mapping (role_id, user_id)
SELECT kr.id, ue.id
FROM keycloak_role kr
JOIN realm r ON r.id = kr.realm_id AND r.name = 'master'
JOIN user_entity ue ON ue.username = 'sysadmin' AND ue.realm_id = r.id
WHERE kr.name = 'sysadmin';