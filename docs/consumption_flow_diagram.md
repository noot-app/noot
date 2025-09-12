# Complete `/api/v1/consumption` POST Flow Diagram

This diagram depicts the entire control flow from audio/text input to consumption storage in the database, including all input methods, caching scenarios, AI provider interactions, and edge cases.

```mermaid
graph TB
    %% Starting point - Multiple input methods
    Start([Input Methods:<br/>1. Audio Upload (multipart)<br/>2. Text Input (JSON/multipart)<br/>3. Consumption Duplication (JSON)]) --> Auth{Dual Authentication<br/>JWT or API Key required}
    
    %% Authentication flow - DUAL AUTHENTICATION REQUIRED
    Auth -->|Valid JWT| AuthJWTSuccess[JWT: Set auth_user in context]
    Auth -->|Valid API Key| AuthAPISuccess[API Key: Set auth_user + scope in context]
    Auth -->|Missing both headers| AuthMissing[Return 401 Authorization required]
    Auth -->|Invalid JWT format| AuthFormat[Return 401 Invalid authorization header format]
    Auth -->|Invalid/expired JWT| AuthInvalid[Return 401 Invalid token or Token expired]
    Auth -->|Invalid API key| AuthAPIInvalid[Return 401 Invalid API key]
    Auth -->|API key revoked/expired| AuthAPIRevoked[Return 401 API key revoked or expired]
    Auth -->|Insufficient API scope| AuthScope[Return 403 Insufficient scope]
    Auth -->|User not found| AuthUser[Return 404 User not found]
    Auth -->|Config error| AuthConfig[Return 500 Authentication configuration error]
    Auth -->|Service unavailable| AuthService[Return 503 Authentication service unavailable]
    
    AuthMissing --> End([End])
    AuthFormat --> End
    AuthInvalid --> End
    AuthAPIInvalid --> End
    AuthAPIRevoked --> End
    AuthScope --> End
    AuthUser --> End
    AuthConfig --> End
    AuthService --> End
    
    %% Request processing - Input normalization
    AuthJWTSuccess --> InputNorm[normalizeConsumptionInput<br/>Detect content type and extract data]
    AuthAPISuccess --> InputNorm
    
    %% Input method detection
    InputNorm --> InputType{Input Method?}
    
    %% JSON consumption duplication
    InputType -->|JSON consumption_id| DuplicateFlow[Consumption Duplication Flow<br/>Skip AI processing]
    
    %% JSON or multipart text input
    InputType -->|JSON text or multipart text| TextFlow[Text Processing Flow<br/>Skip transcription]
    
    %% Multipart audio input
    InputType -->|multipart audio| AudioFlow[Audio Processing Flow<br/>Transcription required]
    
    %% Error handling for invalid input
    InputType -->|Invalid/missing data| InputError[Return 400 Bad Request]
    InputError --> End
    
    %% CONSUMPTION DUPLICATION FLOW
    DuplicateFlow --> ValidateOwnership[Get existing consumption<br/>with ownership check]
    ValidateOwnership -->|Not found/no access| NotFound[Return 404 Consumption not found]
    ValidateOwnership -->|Success| DuplicateConsumption[Create new consumption record<br/>Copy transcript, nutrition, title, note]
    DuplicateConsumption --> SaveItems[Save consumption items<br/>Copy all items with nutrition]
    SaveItems --> CopyLabels[Copy labels from original]
    CopyLabels --> DuplicateResponse[Return new consumption as JSON]
    NotFound --> End
    DuplicateResponse --> End
    
    %% AUDIO PROCESSING FLOW
    AudioFlow --> ParseForm[Parse multipart form<br/>Max 50MB size limit]
    ParseForm -->|Success| ExtractAudio[Extract audio file from audio field]
    ParseForm -->|Error| FormError[Return 400 Bad Request]
    FormError --> End
    
    ExtractAudio -->|Success| SaveTemp[Save to temporary file<br/>with MIME type detection]
    ExtractAudio -->|No audio field| AudioError[Return 400 Bad Request]
    AudioError --> End
    
    %% Phase 1: Transcription (Audio only)
    SaveTemp --> CreateService[Create NutritionService<br/>with store and AI provider]
    CreateService --> Transcribe[TranscribeAudio via OpenAI<br/>Whisper /audio/transcriptions API]
    Transcribe -->|Success| TranscribeSuccess[Sanitized transcript text]
    Transcribe -->|Error| TranscribeError[Return 500 Internal Server Error]
    TranscribeError --> CleanupTemp[Delete temporary file] --> End
    
    %% TEXT PROCESSING FLOW (Direct to Phase 2)
    TextFlow --> CreateServiceText[Create NutritionService<br/>with store and AI provider]
    CreateServiceText --> TextReady[Sanitized text ready]
    
    %% Convergence point for both audio and text
    TranscribeSuccess --> AIProcessing[AI Processing Pipeline]
    TextReady --> AIProcessing
    
    %% Phase 2: Parse Items (Common for audio and text)
    AIProcessing --> ParseItems[ParseItems via OpenAI<br/>Extract food items from text]
    ParseItems -->|Success| ValidateItems[Validate and sanitize item data<br/>Check name length quantities etc]
    ParseItems -->|Error| ParseError[Return 500 Internal Server Error]
    ParseError --> CleanupTempOptional[Delete temporary file if audio]
    CleanupTempOptional --> End
    
    %% Phase 3: Nutrition Hydration - CRITICAL: NEVER FAILS THE REQUEST
    ValidateItems --> CheckStore{Store available?}
    CheckStore -->|Yes| HydrateWithCache[HydrateNutrition with cache<br/>ALWAYS SUCCEEDS even if items fail]
    CheckStore -->|No| HydrateNoCache[HydrateNutritionWithoutCache<br/>ALWAYS SUCCEEDS even if items fail]
    
    %% Parallel processing of items with cache
    HydrateWithCache --> ParallelCache[Process items in parallel goroutines<br/>Individual failures are handled gracefully]
    ParallelCache --> ItemLoop[For each item hydrateItemNutrition]
    
    %% Cache flow for individual items
    ItemLoop --> CacheCheck[fetchNutritionFromCache]
    
    %% Cache hit scenarios - 4-tier caching system
    CacheCheck --> CanonicalCache{Canonical food cache hit<br/>Deduplication by canonical name<br/>Fresh within 30 days?}
    CanonicalCache -->|Yes Fresh| CanonicalHit[Use canonical cached data<br/>Scale using per-100g nutrition<br/>Handle BaseQuantity scaling]
    CanonicalCache -->|No| ExactCache{Exact serving cache hit<br/>Exact name + brand + grams<br/>Fresh within 30 days?}
    ExactCache -->|Yes Fresh| ExactHit[Use cached exact serving data<br/>Scale if BaseQuantity greater than 1]
    ExactCache -->|No| FallbackCache{Fallback cache key hit<br/>Backward compatibility<br/>Fresh within 30 days?}
    FallbackCache -->|Yes Fresh| FallbackHit[Use fallback cached data]
    FallbackCache -->|No| ScalableCache{Scalable serving cache hit<br/>Scale from different serving sizes<br/>Fresh within 30 days?}
    ScalableCache -->|Yes Fresh| ScaleHit[Scale nutrition using per-100g data<br/>or serving-based scaling]
    ScalableCache -->|No| CacheMiss[Cache miss fetch from AI]
    
    %% Cache hits go directly to result
    CanonicalHit --> ItemComplete[Item hydrated with nutrition]
    ExactHit --> ItemComplete
    FallbackHit --> ItemComplete
    ScaleHit --> ItemComplete
    
    %% Cache miss flow - Always uses complete AI response
    CacheMiss --> FetchContext[Check if item has brand<br/>for Open Food Facts context]
    FetchContext --> OFFCheck{Item has brand?}
    
    %% OFF Open Food Facts flow
    OFFCheck -->|Yes| OFFQuery[Query OFF database<br/>SearchProduct by name plus brand]
    OFFCheck -->|No| NoOFF[Skip OFF query]
    
    OFFQuery -->|Found| OFFFound[Extract ingredients URL<br/>Create nutrition context for AI]
    OFFQuery -->|Not found| OFFNotFound[No OFF context]
    
    OFFFound --> FetchAI[fetchNutritionFromAI]
    OFFNotFound --> FetchAI
    NoOFF --> FetchAI
    
    %% AI Provider flow - Always uses complete response
    FetchAI --> CompleteAI[GetNutritionWithContextComplete<br/>Always get nutrition + ingredients + URL<br/>for both generic and branded items]
    
    CompleteAI --> AISuccess[AI response with complete data<br/>includes nutrition, ingredients, URL]
    FetchAI -->|Error| AIError[Log warning and use item without nutrition<br/>DOES NOT FAIL THE REQUEST]
    
    AISuccess --> CacheStore[cacheNutritionData<br/>Store using canonical food names<br/>for deduplication in items table]
    CacheStore --> ItemComplete
    AIError --> ItemWithoutNutrition[Item without nutrition data<br/>Will get Note field in response]
    
    %% Parallel processing without cache
    HydrateNoCache --> ParallelNoCache[Process items in parallel goroutines<br/>Individual failures handled gracefully]
    ParallelNoCache --> DirectOFF[Direct OFF query if branded]
    DirectOFF --> DirectAI[Direct AI provider calls]
    DirectAI --> ItemCompleteNoCache[Item hydrated or without nutrition]
    
    %% Collect results - ALWAYS SUCCEEDS
    ItemComplete --> CollectResults[Collect all parallel results<br/>HYDRATION ALWAYS SUCCEEDS]
    ItemWithoutNutrition --> CollectResults
    ItemCompleteNoCache --> CollectResults
    
    %% Phase 4: Response preparation
    CollectResults --> ConvertAPI[Convert internal types to API types<br/>Items without nutrition get Note field]
    ConvertAPI --> Summarize[Calculate nutrition summary<br/>totals for all items]
    
    %% Phase 5: Database storage
    Summarize --> StoreCheck{Store available?}
    StoreCheck -->|Yes| GetUser[getCurrentUser from JWT context]
    StoreCheck -->|No| SkipStorage[Skip database storage]
    
    GetUser -->|Success| CreateConsumption[Create Consumption record<br/>with transcript and totals]
    GetUser -->|Error| LogUserError[Log error continue without storage<br/>DOES NOT FAIL REQUEST]
    
    CreateConsumption -->|Success| SaveConsumption[CreateConsumption in database]
    CreateConsumption -->|Error| LogCreateError[Log error continue<br/>DOES NOT FAIL REQUEST]
    
    SaveConsumption -->|Success| SaveItems[Create ConsumptionItem records<br/>for each individual item]
    SaveConsumption -->|Error| LogSaveError[Log error continue<br/>DOES NOT FAIL REQUEST]
    
    SaveItems --> LinkItems[Try to link items to global cache<br/>using normalized names<br/>Individual failures logged but ignored]
    
    %% Response generation - ALWAYS SUCCEEDS
    LinkItems --> CreateResponse[Create ConsumptionResponse<br/>Convert to API format]
    SkipStorage --> CreateResponse
    LogUserError --> CreateResponse
    LogCreateError --> CreateResponse
    LogSaveError --> CreateResponse
    
    CreateResponse --> ResponseData[Response includes<br/>Consumption ID Transcript<br/>Items with nutrition or Note field<br/>48+ nutrition totals Request ID<br/>Labels and metadata]
    
    %% Final cleanup and response
    ResponseData --> CleanupTemp2[Delete temporary audio file if present]
    CleanupTemp2 --> Success[Return 200 OK with JSON<br/>ALWAYS SUCCEEDS after auth]
    Success --> End
    
    %% Styling - High contrast for accessibility
    classDef errorClass fill:#B71C1C,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef successClass fill:#1B5E20,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef cacheClass fill:#0D47A1,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef aiClass fill:#E65100,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef dbClass fill:#4A148C,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef authClass fill:#BF360C,stroke:#000000,stroke-width:3px,color:#FFFFFF
    classDef inputClass fill:#6A1B9A,stroke:#000000,stroke-width:3px,color:#FFFFFF
    
    class AuthMissing,AuthFormat,AuthInvalid,AuthAPIInvalid,AuthAPIRevoked,AuthScope,AuthUser,AuthConfig,AuthService,FormError,AudioError,TranscribeError,ParseError,InputError,NotFound errorClass
    class Success,CanonicalHit,ExactHit,FallbackHit,ScaleHit,DuplicateResponse successClass
    class CacheCheck,CanonicalCache,ExactCache,FallbackCache,ScalableCache,CacheStore cacheClass
    class Transcribe,ParseItems,CompleteAI aiClass
    class CreateConsumption,SaveConsumption,SaveItems,DuplicateConsumption dbClass
    class Auth,AuthJWTSuccess,AuthAPISuccess authClass
    class InputNorm,DuplicateFlow,TextFlow,AudioFlow inputClass
```

## Key Edge Cases and Flows Covered

### 1. Dual Authentication Edge Cases (REQUIRED FOR CONSUMPTION ENDPOINT)

**JWT Authentication:**
- Missing Authorization header (when no X-API-Key) → 401 "Authorization required"
- Invalid authorization header format → 401 "Invalid authorization header format"  
- Invalid JWT token → 401 "Invalid token"
- Expired JWT token → 401 "Token expired"
- Valid token but user not found → 404 "User not found"
- JWKS/key resolution issues → 503 "Authentication service temporarily unavailable"
- Configuration errors → 500 "Authentication configuration error"

**API Key Authentication (Pro users only):**
- Invalid API key format → 401 "Invalid API key"
- API key not found → 401 "Invalid API key"
- API key revoked or expired → 401 "API key revoked or expired"
- Insufficient scope (read-only key on write operation) → 403 "Insufficient scope"
- User downgraded from Pro → 403 "API key requires Pro subscription"
- API key lookup failures → 503 "Authentication service temporarily unavailable"

### 2. Input Method Edge Cases

**Consumption Duplication (JSON):**
- Invalid consumption_id → 404 "Consumption not found"
- Access denied (not owner) → 404 "Consumption not found or access denied"
- Database errors → 500 Internal Server Error

**Text Input (JSON or multipart):**
- Empty text field → 400 "Text field cannot be empty"
- Text sanitization removes all content → 400 "Text field cannot be empty"
- Invalid JSON format → 400 "Invalid JSON payload"

**Audio Upload (multipart):**
- Invalid multipart form → 400 Bad Request
- Missing audio field (when no text) → 400 Bad Request
- File too large (>50MB) → 400 Bad Request
- Successful upload → Temporary file creation with cleanup

### 3. Four-Tier Cache Hit Scenarios (30-day TTL)

- **Canonical food cache hit**: Deduplication using canonical food names, scales using per-100g nutrition (fresh within 30 days)
- **Exact serving cache hit**: Direct match on normalized name + brand + grams (fresh within 30 days)
- **Fallback cache hit**: Backward compatibility with old cache keys (fresh within 30 days)
- **Scalable serving cache hit**: Scale nutrition from different serving sizes using per-100g or serving-based scaling (fresh within 30 days)
- **Cache miss**: Proceeds to AI provider with Open Food Facts context

### 4. AI Provider Integration (NEVER FAILS THE REQUEST)

- **All items**: Always use `GetNutritionWithContextComplete` to get nutrition + ingredients + URL
- **Branded items**: Enhanced with Open Food Facts context data when available
- **Generic items**: Direct AI analysis with comprehensive nutrition breakdown
- **AI failures**: Log warnings and continue with items without nutrition data - NEVER fails entire request
- **Transcription failures**: Return 500 error (only phase that can fail the request)
- **Parsing failures**: Return 500 error (only phase that can fail the request)

### 5. Open Food Facts (OFF) Integration

- Only query OFF for items with brands (branded products)
- Extract ingredients, URLs, and nutrition context for AI
- Fallback gracefully if OFF query fails

### 6. Database Storage Edge Cases (NEVER FAIL THE REQUEST)

- **No store available**: Skip all database operations, return response
- **User authentication fails**: Log error, continue without storage
- **Consumption creation fails**: Log error, continue
- **Individual item storage fails**: Log error, continue with other items
- **Cache storage fails**: Log warning, continue (don't fail response)

### 7. Parallel Processing (ROBUST ERROR HANDLING)

- Process multiple food items concurrently using goroutines
- **Individual item failures are handled gracefully** - they get items without nutrition
- **HydrateNutrition ALWAYS succeeds** - even if all individual items fail
- Collect results and continue even if some items fail nutrition lookup

### 8. Data Validation and Security

- **Input sanitization**: Sanitize all text inputs (transcript, user text) using bluemonday
- **Item validation**: Validate item names, quantities, and brand lengths
- **Authentication security**: Dual auth with scope validation for API keys
- **Database security**: Row Level Security (RLS) policies enforce user data isolation
- **File security**: Temporary file handling with cleanup, MIME type detection
- **Request limits**: 50MB upload limit, timeout configurations

### 9. Critical Success Guarantee

**After successful authentication, most requests ALWAYS return 200 OK**:

**Request-failing errors (500):**
- Transcription failures (audio input only)
- Item parsing failures (audio and text input)

**Gracefully handled errors (logged but don't fail request):**
- Nutrition hydration failures → Items get "Note" field: "Nutrition data unavailable"
- Database storage failures → Response returned without storage
- Cache storage failures → Response returned without caching
- Individual item failures → Other items continue processing
- Label copying failures → Consumption created without labels

**Response always includes:**
- Consumption ID and metadata
- Original transcript or text
- Items with nutrition data or Note field
- Complete nutrition summary (48+ nutrients)
- Request ID for tracking
