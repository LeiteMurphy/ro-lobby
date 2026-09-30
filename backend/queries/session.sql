-- name: SessionTimeZone :one
-- Fuso da sessão atual, usado para conferir a RN-08.
SELECT current_setting('TimeZone')::text AS time_zone;
