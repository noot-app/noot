# Database Security Model

## Overview

The Noot application implements Supabase's recommended security pattern for managing auth and profile data. This document explains our security model and why **we do not need to grant DML permissions to `supabase_auth_admin` on `public.users`**.

## Architecture

### Current Implementation ✅

Our setup follows the recommended pattern:

1. **Auth Data**: Managed by Supabase Auth in the `auth.users` table
2. **Profile Data**: Stored in `public.users` with a foreign key reference to `auth.users(id)`
3. **Automatic Profile Creation**: Database trigger creates profile when auth user is created
4. **Row Level Security**: RLS policies protect access to user data

### Why This is Secure

#### 1. Separation of Concerns
- **`auth.users`**: Managed exclusively by Supabase Auth
- **`public.users`**: App-specific profile data with controlled access

#### 2. Trigger-Based Profile Creation
```sql
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
```

The trigger function uses `SECURITY DEFINER`, which means it runs with the privileges of the function owner (superuser), not the caller. This provides the necessary elevated privileges for creating profiles without granting broad permissions to `supabase_auth_admin`.

#### 3. Row Level Security Protection
```sql
-- Users can only view their own profile
CREATE POLICY "Users can view own profile" ON users
  FOR SELECT USING (auth.uid() = id);

-- Users can update their own profile (but not subscription_tier)
CREATE POLICY "Users can update own profile" ON users  
  FOR UPDATE USING (auth.uid() = id)
  WITH CHECK (
    auth.uid() = id AND
    subscription_tier = (SELECT subscription_tier FROM users WHERE id = auth.uid()) AND
    id = auth.uid()
  );

-- System can insert users (via trigger)
CREATE POLICY "System can insert users" ON users
  FOR INSERT WITH CHECK (true);
```

## Why We Don't Need `supabase_auth_admin` Grants

### The Problem with Granting DML Permissions

If we were to grant permissions like:
```sql
GRANT SELECT, INSERT, UPDATE, DELETE ON public.users TO supabase_auth_admin;
```

This would create security risks:

1. **Overprivileged Access**: The auth system would have full DML access to our app data
2. **Trust Boundary Violation**: Mixing auth system privileges with app data access
3. **Uncontrolled Modifications**: Auth system could modify or delete app-specific data

### Our Secure Alternative

Instead of granting broad permissions, we use:

1. **`SECURITY DEFINER` trigger function**: Provides controlled, specific access for profile creation
2. **RLS policies**: Ensure users can only access their own data
3. **Foreign key constraints**: Maintain data integrity between auth and profile tables

## Common Issues and Solutions

### "Database error updating user" During Signup

If you encounter this error, it's typically caused by:

1. **Missing trigger**: Ensure the `on_auth_user_created` trigger exists
2. **Function permissions**: Verify the trigger function has `SECURITY DEFINER`
3. **RLS policy conflicts**: Check that the "System can insert users" policy exists

**Solution**: Our current setup addresses all these issues correctly.

### Profile Data Access

- **Frontend**: Use RLS-protected queries via Supabase client
- **Backend**: Use service role key for administrative operations
- **Never**: Grant DML permissions to `supabase_auth_admin`

## Verification

To verify our setup is working correctly:

1. **Check trigger exists**:
   ```sql
   SELECT * FROM information_schema.triggers 
   WHERE trigger_name = 'on_auth_user_created';
   ```

2. **Test profile creation**: Sign up a new user and verify profile is created automatically

3. **Verify RLS policies**:
   ```sql
   SELECT * FROM pg_policies WHERE tablename = 'users';
   ```

## Best Practices

1. ✅ **Use triggers** for automatic profile creation
2. ✅ **Enable RLS** on all app tables
3. ✅ **Use `SECURITY DEFINER`** for system operations
4. ❌ **Never grant** DML permissions to `supabase_auth_admin`
5. ❌ **Don't bypass** RLS with overprivileged roles

## References

- [Supabase: Managing User Data](https://supabase.com/docs/guides/auth/managing-user-data)
- [PostgreSQL: Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [PostgreSQL: Security Definer Functions](https://www.postgresql.org/docs/current/sql-createfunction.html#SQL-CREATEFUNCTION-SECURITY)