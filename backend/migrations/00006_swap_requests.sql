-- Pedidos de troca de personagem e bloqueio na remoção (spec candidatura-lobby, Parte 2,
-- design D-08 e D-09). Os CHECK repetem as regras do serviço como última defesa.

-- +goose Up
-- RN-15: só fica true numa candidatura removida com bloqueio.
ALTER TABLE applications ADD COLUMN blocked boolean NOT NULL DEFAULT false;

CREATE TABLE swap_requests (
    id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- O membro e o lobby vêm da candidatura aceita.
    application_id    uuid        NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    from_character_id uuid        REFERENCES characters (id) ON DELETE SET NULL,
    -- RN-20: personagem do próprio membro; RN-26 trava a exclusão enquanto pendente.
    to_character_id   uuid        REFERENCES characters (id) ON DELETE SET NULL,
    to_role           text        NOT NULL CHECK (to_role IN ('tank', 'support', 'dps')),
    -- RN-20: motivo de 10 a 250 caracteres.
    reason            text        NOT NULL CHECK (char_length(btrim(reason)) BETWEEN 10 AND 250),
    status            text        NOT NULL CHECK (status IN (
                                      'pending', 'accepted', 'rejected', 'withdrawn',
                                      'expired', 'cancelled')),                        -- RN-24
    -- RN-22: justificativa da recusa, de 10 a 250 caracteres.
    decision_reason   text        CHECK (char_length(btrim(decision_reason)) BETWEEN 10 AND 250),
    created_at        timestamptz NOT NULL,
    decided_at        timestamptz
);

-- RN-20: no máximo um pedido pendente por membro e lobby.
CREATE UNIQUE INDEX swap_requests_one_pending
    ON swap_requests (application_id) WHERE status = 'pending';
CREATE INDEX swap_requests_to_character_pending_idx
    ON swap_requests (to_character_id) WHERE status = 'pending';

-- RN-18: cada transição do pedido, como em application_events.
CREATE TABLE swap_request_events (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    swap_request_id uuid        NOT NULL REFERENCES swap_requests (id) ON DELETE CASCADE,
    from_status     text,
    to_status       text        NOT NULL,
    actor_id        uuid        REFERENCES users (id) ON DELETE SET NULL,
    reason          text,
    at              timestamptz NOT NULL
);
CREATE INDEX swap_request_events_swap_request_id_idx ON swap_request_events (swap_request_id);

-- +goose Down
DROP TABLE swap_request_events;
DROP TABLE swap_requests;
ALTER TABLE applications DROP COLUMN blocked;
