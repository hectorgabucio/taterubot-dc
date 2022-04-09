CREATE TABLE IF NOT EXISTS public.voicedata (
                                  id varchar NOT NULL,
                                  guildid varchar NOT NULL,
                                  "timestamp" timestamptz NOT NULL,
                                  "name" varchar NOT NULL,
                                  userid varchar NOT NULL,
                                  duration numeric NOT NULL,
                                  CONSTRAINT voicedata_pk PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS voicedata_guildid_idx ON public.voicedata USING btree (guildid);
CREATE INDEX IF NOT EXISTS voicedata_timestamp_idx ON public.voicedata USING btree ("timestamp");