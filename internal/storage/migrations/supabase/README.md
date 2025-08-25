# Notes

These files came over from the `apps/web/` directory which was originally a svelte template that said it supported supabase. I copied these migration files over because the stuff around stripe, avatars, and contact_requests all seemed pretty useful. The issue right now is I have a local sqlite db for development that has a users table and in production, I want my entire postgres db and all the users stuff (like auth) to live in supabase. So I need a good way to stitch this all together.

I have created a single user in supabase by doing the "sign up flow" in my frontend via local development and it resulted in a data structure that looks like the contents of this file next to this readme: `internal/storage/migrations/supabase/supabase-user-auth-db-structure.json`.

Helpful docs:

- https://supabase.com/docs/guides/auth/jwts
- https://supabase.com/docs/guides/auth/signing-keys
- https://supabase.com/blog/jwt-signing-keys
