# Security Guidelines

This document outlines security best practices and guidelines for the noot application, based on the comprehensive security review conducted.

## Security Review Summary

A thorough security review was conducted covering:
- ✅ Authentication & Authorization
- ✅ Data Access & Leakage Protection  
- ✅ SQL Injection Prevention
- ✅ Input Validation & Sanitization
- ✅ Session & Token Handling

**Result: No critical security flaws identified.**

## Key Security Features

### 1. Authentication & Authorization
- **JWT Validation**: Proper Supabase JWT token validation with JWKS support
- **User Context**: All API requests validate user identity and create secure user context
- **Authorization Checks**: Subscription tier validation (`RequireProSubscription`)
- **Row-Level Security (RLS)**: Database-level access control for all user data

### 2. Data Protection
- **Parameterized Queries**: All SQL queries use parameters to prevent injection
- **RLS Policies**: Users can only access their own data (profiles, consumptions, etc.)
- **No Data Leakage**: API responses filtered, no sensitive fields exposed in logs

### 3. Input Validation
- **AI Output Validation**: All AI responses are validated and sanitized
- **Length Limits**: Input length restrictions prevent abuse
- **Content Filtering**: Dangerous patterns detected and filtered from user input

## AI/LLM Security Measures

Since this application processes user voice input through AI services, special security measures are implemented:

### Input Validation
```go
// Maximum lengths for security
const (
    maxTranscriptLength = 10000  // Prevent oversized transcripts
    maxItemNameLength   = 200    // Limit food item names  
    maxBrandLength      = 100    // Limit brand names
)
```

### Output Sanitization
- Control character removal
- SQL injection pattern detection
- XSS pattern filtering
- Input length validation

### Secure Logging
- Transcript data is truncated in logs
- No sensitive user data exposed in debug output
- Structured error responses without internal details

## Development Security Guidelines

### Adding New API Endpoints

When adding new API endpoints, **ALWAYS**:

1. **Use Authentication Middleware**
   ```go
   // Apply JWT authentication
   api.Use(JWTAuthMiddleware(store))
   
   // For protected routes
   protected := api.Group("/")
   protected.Use(RequireAuth())
   
   // For pro features  
   pro := api.Group("/")
   pro.Use(RequireProSubscription())
   ```

2. **Validate User Context**
   ```go
   func (s *APIServer) MyEndpoint(c *gin.Context) {
       user := GetAuthenticatedUser(c)
       if user == nil {
           // This should not happen if middleware is correct
           c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
           return
       }
       // Use user.ID for all database operations
   }
   ```

3. **Use Parameterized Queries**
   ```go
   // CORRECT
   err := s.store.GetUserData(ctx, user.ID, itemID)
   
   // NEVER DO THIS
   query := fmt.Sprintf("SELECT * FROM items WHERE user_id = %s", userID)
   ```

4. **Validate Input Data**
   ```go
   // Validate all user input
   if len(request.Name) > maxItemNameLength {
       return NewAppError("Name too long", http.StatusBadRequest, nil)
   }
   
   // Sanitize text fields
   request.Name = sanitizeText(request.Name)
   ```

### Database Changes

When modifying database schema:

1. **Enable RLS on all user tables**
   ```sql
   alter table public.my_table enable row level security;
   ```

2. **Create appropriate RLS policies**
   ```sql
   create policy "Users can only access own data"
   on public.my_table
   for all using (auth.uid() = user_id);
   ```

3. **Test access controls**
   - Verify users cannot access other users' data
   - Test with different subscription tiers

### AI/LLM Integration

When working with AI services:

1. **Validate all inputs before sending to AI**
   ```go
   if len(userInput) > maxInputLength {
       return errors.New("input too long")
   }
   ```

2. **Sanitize all AI outputs before use**
   ```go
   result := sanitizeAIOutput(aiResponse)
   validatedResult := validateStructuredData(result)
   ```

3. **Use structured prompts**
   - Define clear expected output formats
   - Validate response structure
   - Handle malformed responses gracefully

4. **Log safely**
   ```go
   // DON'T log full user input
   LogDebug("Processing request", "input", userInput)
   
   // DO log truncated preview
   LogDebug("Processing request", "input_preview", truncateForLog(userInput))
   ```

## Security Monitoring

### Logging Best Practices
- Never log JWT tokens or passwords
- Truncate user inputs in logs
- Use structured logging with appropriate levels
- Log authentication failures for monitoring

### Error Handling
- Return generic error messages to users
- Log detailed errors server-side only
- Use appropriate HTTP status codes
- Include request IDs for tracing

## Regular Security Tasks

1. **Dependency Updates**: Regularly update dependencies to patch vulnerabilities
2. **JWT Secret Rotation**: Plan for JWT secret rotation in production
3. **Access Log Review**: Monitor authentication failures and unusual patterns
4. **RLS Policy Review**: Verify RLS policies when adding new database features
5. **AI Output Validation**: Review AI output validation rules as models change

## Security Testing

Before deploying changes:

1. **Test Authentication**
   - Verify unauthenticated requests are rejected
   - Test token expiration handling
   - Validate user context isolation

2. **Test Authorization** 
   - Verify subscription tier restrictions
   - Test cross-user data access prevention
   - Validate RLS policies

3. **Test Input Validation**
   - Try oversized inputs
   - Test special characters and control sequences
   - Validate AI output sanitization

4. **Review Logs**
   - Ensure no sensitive data in logs
   - Verify error messages don't leak information
   - Check authentication failure logging

## Incident Response

If a security issue is discovered:

1. **Immediate**: Stop processing if data exposure risk
2. **Assess**: Determine scope and impact
3. **Mitigate**: Apply immediate fixes
4. **Review**: Audit logs for compromise
5. **Document**: Update security guidelines
6. **Test**: Verify fix effectiveness

## Contact

For security concerns or questions about these guidelines, please contact the development team or create a security-related issue in the repository.