CREATE DATABASE testcoins
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;

-- accounts table
CREATE TABLE IF NOT EXISTS public.accounts
(
    account_number bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    balance bigint NOT NULL DEFAULT 0,
    CONSTRAINT accounts_pkey PRIMARY KEY (account_number)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.accounts
    OWNER to postgres;

INSERT INTO public.accounts(amount)
	VALUES (10000), (250000);


-- trasnfers table
CREATE TABLE IF NOT EXISTS public.transfers
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    transfer_id uuid NOT NULL,
    source_account bigint NOT NULL,
    dest_account bigint NOT NULL,
    amount bigint NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT transfers_pkey PRIMARY KEY (id),
    CONSTRAINT transfers_accounts_dest_fkey FOREIGN KEY (dest_account)
        REFERENCES public.accounts (account_number) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT transfers_accounts_source_fkey FOREIGN KEY (source_account)
        REFERENCES public.accounts (account_number) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.transfers
    OWNER to postgres;

-- Index: idx_dest_account
CREATE INDEX IF NOT EXISTS idx_dest_account
    ON public.transfers USING btree
    (dest_account ASC NULLS LAST)
    TABLESPACE pg_default;

-- Index: idx_source_account
CREATE INDEX IF NOT EXISTS idx_source_account
    ON public.transfers USING btree
    (source_account ASC NULLS LAST)
    TABLESPACE pg_default;

-- Index: idx_transfers_transaction_id
CREATE INDEX IF NOT EXISTS idx_transfers_transaction_id
    ON public.transfers USING btree
    (transfer_id ASC NULLS LAST)
    INCLUDE(transfer_id)
    TABLESPACE pg_default;