package model

import (
	"net/url"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. Basic personal info
type BasicPersonInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Age  *int   `json:"age,omitempty"`
}

// 2. Nested struct with pointer, slice, and map
type NestedBasicInfo struct {
	BasicInfo *BasicPersonInfo       `json:"basic_info"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// 3. Embedded struct field
type EmbeddedBasicInfo struct {
	BasicPersonInfo
	ExtraField string `json:"extra_field"`
}

// 4. Type alias for custom int
type CustomInt int

type AliasWithCustomInt struct {
	CustomValue CustomInt `json:"custom_value"`
}

// 5. Struct with timestamp fields
type WithTimestamps struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// 6. Struct with complex nested collections
type ComplexNestedCollections struct {
	Items       []NestedBasicInfo             `json:"items"`
	Attributes  map[string][]*BasicPersonInfo `json:"attributes,omitempty"`
	Flags       map[string]bool               `json:"flags"`
	OptionalPtr *EmbeddedBasicInfo            `json:"optional_ptr,omitempty"`
	Aliases     map[CustomInt]string          `json:"aliases"`
}

// 7. Anonymous embedded pointer field
type AnonymousEmbeddedBasic struct {
	*BasicPersonInfo
	Score float64 `json:"score"`
}

// 8. Struct with any type field
type WithAnyData struct {
	Data interface{} `json:"data"`
}

// 9. Enum for user status
type UserStatus int

const (
	StatusActive UserStatus = iota
	StatusInactive
	StatusSuspended
)

// 10. Email type alias
type Email string

// 11. User account
type UserAccount struct {
	ID          int                    `json:"id"`
	Email       Email                  `json:"email"`
	Name        string                 `json:"name"`
	Status      UserStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
	Profile     *UserProfileDetail     `json:"profile,omitempty"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// 12. User profile detail
type UserProfileDetail struct {
	Nickname    string            `json:"nickname"`
	Birthday    *time.Time        `json:"birthday,omitempty"`
	Website     *url.URL          `json:"website,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
}

// 13. Admin account
type AdminAccount struct {
	UserAccount
	AdminLevel int `json:"admin_level"`
}

// 14. Polymorphic payload
type Payload interface{}

type MessageWithPayload struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   Payload   `json:"payload"`
}

// 15. User cache
type UserCache struct {
	Items map[string][]*UserAccount `json:"items"`
	TTL   int                       `json:"ttl_seconds"`
}

// 16. Complex profile collection
type ComplexProfileCollections struct {
	Values       []*UserProfileDetail          `json:"values"`
	Lookup       map[string]*UserProfileDetail `json:"lookup"`
	ExtraOptions map[string]map[int]string     `json:"extra_options"`
}

// 17. Inline example
type InlineBasicExample struct {
	ID     int             `json:"id"`
	Inline BasicPersonInfo `json:",inline"`
	Notes  string          `json:"notes,omitempty"`
}

// 18. Product category
type ProductCategory struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ParentID    *int       `json:"parent_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// 19. Store item
type StoreItem struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	CategoryID  int               `json:"category_id"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	Stock       int               `json:"stock"`
	IsActive    bool              `json:"is_active"`
	Tags        []string          `json:"tags,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
}

// 20. Sales order
type SalesOrder struct {
	ID              int              `json:"id"`
	UserID          int              `json:"user_id"`
	OrderItems      []SalesOrderItem `json:"order_items"`
	TotalPrice      float64          `json:"total_price"`
	Currency        string           `json:"currency"`
	Status          OrderStatus      `json:"status"`
	OrderedAt       time.Time        `json:"ordered_at"`
	DeliveredAt     *time.Time       `json:"delivered_at,omitempty"`
	ShippingAddress string           `json:"shipping_address,omitempty"`
	PaymentMethod   PaymentMethod    `json:"payment_method"`
}

// 21. Sales order item
type SalesOrderItem struct {
	ItemID    int     `json:"item_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// 22. Order status enum
type OrderStatus int

const (
	OrderPending OrderStatus = iota
	OrderProcessing
	OrderShipped
	OrderDelivered
	OrderCancelled
)

// 23. Payment method enum
type PaymentMethod int

const (
	PaymentCreditCard PaymentMethod = iota
	PaymentPayPal
	PaymentToss
	PaymentOther
)

// 24. MongoDB data model
type MongoDBDataModel struct {
	ID        primitive.ObjectID `json:"id"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at,omitempty"`
}

// 25. All primitive types
type AllPrimitiveTypes struct {
	IntVal        int        `json:"int_val"`
	Int8Val       int8       `json:"int8_val"`
	Int16Val      int16      `json:"int16_val"`
	Int32Val      int32      `json:"int32_val"`
	Int64Val      int64      `json:"int64_val"`
	UintVal       uint       `json:"uint_val"`
	Uint8Val      uint8      `json:"uint8_val"`
	Uint16Val     uint16     `json:"uint16_val"`
	Uint32Val     uint32     `json:"uint32_val"`
	Uint64Val     uint64     `json:"uint64_val"`
	Float32Val    float32    `json:"float32_val"`
	Float64Val    float64    `json:"float64_val"`
	BoolVal       bool       `json:"bool_val"`
	ByteVal       byte       `json:"byte_val"`
	RuneVal       rune       `json:"rune_val"`
	Complex64Val  complex64  `json:"complex64_val"`
	Complex128Val complex128 `json:"complex128_val"`
	StringVal     string     `json:"string_val"`
	BytesSlice    []byte     `json:"bytes_slice"`
}

// 26. DB-specific type examples
type DBSpecificTypes struct {
	ObjectIDField primitive.ObjectID `json:"object_id_field"`
	Decimal128    interface{}        `json:"decimal_128"`
	UUIDField     string             `json:"uuid_field"`
}

// 27. Composite types
type CompositeFieldTypes struct {
	PtrInt       *int                   `json:"ptr_int"`
	SliceStrings []string               `json:"slice_strings"`
	MapStringAny map[string]interface{} `json:"map_string_any"`
	MapStringInt map[string]int         `json:"map_string_int"`
	MapIntString map[int]string         `json:"map_int_string"`
}

// 28. Generic result wrapper
type GenericResult[T any] struct {
	Data T      `json:"data"`
	Err  string `json:"err,omitempty"`
}

// 29. Event payload with method
type Event interface {
	Process() error
}

type EventWithAnyData struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
}

// 30. Job and worker
type Job struct {
	ID   string
	Name string
}

type JobWorker struct {
	JobChan chan Job
	Stop    func() error
}

// 31. Container with meta
type ContainerWithMeta struct {
	Meta struct {
		Created time.Time `json:"created"`
		Count   int       `json:"count"`
	} `json:"meta"`
}

// 32. Structs with field conflict
type StructAWithField struct {
	Field int `json:"field"`
}

type StructBWithConflict struct {
	StructAWithField
	Field string `json:"field"`
}

// 33. Custom marshal struct
type UserWithCustomMarshal struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (u UserWithCustomMarshal) MarshalJSON() ([]byte, error) {
	return []byte(`{"name": "` + u.Name + `"}`), nil
}

// 34. Self-referencing recursive
type RecursiveNode struct {
	Parent *RecursiveNode   `json:"parent,omitempty"`
	Child  []*RecursiveNode `json:"child,omitempty"`
}

// 35. Unsafe pointer alias
type CPointerAlias unsafe.Pointer

// 36. Generic pair
type GenericPair[A any, B int] struct {
	First  A `json:"first"`
	Second B `json:"second"`
}

// 37. Logger and service
type Logger struct{}

func (l Logger) Log(msg string) {}

type LoggedService struct {
	Logger
	Name string `json:"name"`
}

// 38. Function handler
type FunctionHandler struct {
	Process func(input string) (output string, err error)
	Init    func()
}

// 39. Alias chain
type CustomString string
type UserID CustomString

// 40. Complex slice
type ResultUserList struct {
	Elements []*GenericResult[*UserAccount] `json:"elements"`
}

// 41. Multi-return func holder
type MultipleReturnFunctionHolder struct {
	Compute func(x, y int) (int, error)
}

// 42. Map with custom key
type MapWithUserIDKey struct {
	Data map[UserID]string `json:"data"`
}

// 43. Map with struct key
type MapWithStructPosKey struct {
	KeyData map[struct{ X, Y int }]string
}

// 44. Map with struct key -> number
type MapWithStructStringKeyNumber struct {
	KeyData map[struct{ X, Y string }]int
}

// 45. Generic MyStruct
type GenericMyStruct[T any] struct {
	Field1 int `json:"field1"`
	Field2 string
	Field3 string
	EmbeddedType
}

type EmbeddedType struct {
	EmbField string
}

type AliasTypeMap = map[string]int
type InterfaceAlias interface{}
type EmptyStruct struct{}
type EmbeddedOnly struct {
	*EmbeddedType
}

// 46. New basic types
type NewBasicTypesExample struct {
	StringField string
	IntField    int
	BoolField   bool
	FloatField  float64
	TimeField   time.Time
	URLField    url.URL
	UnsafeField unsafe.Pointer
	Interface   interface{}
	ByteSlice   []byte
	Pointer     *int
	Slice       []string
	MapField    map[string]int
	StructField struct {
		InnerField string
	}
	GenericField T
}

// type parameter
type T any
type AliasIntType = int
type AliasMapType = map[string]string

// 47. Product test item
type ProductTestItem struct {
	ID          primitive.ObjectID     `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Status      int                    `json:"status"`
	Price       ItemPriceInfo          `json:"price,omitempty"`
	Sale        *ItemSaleInfo          `json:"sale,omitempty"`
}

type ItemSalePrice struct {
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
}

type ItemSaleInfo struct {
	Name      string          `json:"name,omitempty"`
	StartedAt *time.Time      `json:"started_at,omitempty"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	Prices    []ItemSalePrice `json:"prices"`
}

type ItemPriceInfo struct {
	Usd float64 `json:"usd,omitempty"`
	Krw float64 `json:"krw,omitempty"`
}

// 48. Postgres model
type PostgresDataModel struct {
	ID        uuid.UUID              `json:"id"`
	JsonData  map[string]interface{} `json:"json_data"`
	IntArray  []int                  `json:"int_array"`
	TextArray []string               `json:"text_array"`
	CreatedAt time.Time              `json:"created_at"`
}

// 49. Mongo test model
type MongoTestDataModel struct {
	ID        primitive.ObjectID     `json:"id"`
	Profile   map[string]interface{} `json:"profile,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

// 50. Redis cache entry
type RedisCacheEntry struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastAccess time.Time `json:"last_access"`
}

// 51. Nested generic
type ResultWithUserList struct {
	Data GenericResult[[]*UserAccount] `json:"data"`
}

// 52. Multi-dimensional
type UserMatrix struct {
	Matrix [][]*BasicPersonInfo `json:"matrix"`
}

type StringMapAliasType = map[string]string

// 53. Uses alias map
type FieldUsingStringMapAlias struct {
	Field StringMapAliasType `json:"field"`
}

// 54. Channel example
type StringChannelExample struct {
	EventChan chan string `json:"event_chan"`
}

// 55. Function example
type TransformFunctionExample struct {
	Transformer func(int) string `json:"transformer"`
}

// 56. Unsafe example
type UnsafePointerExample struct {
	Data unsafe.Pointer `json:"data"`
}

// 57. No tag fields
type NoJSONTagFields struct {
	Field1 string
	Field2 int
}

// 58. Empty type
type EmptyAnyFieldType struct {
	Unknown interface{}
}

// 59. Map key combination
type MapWithCustomAndStructKeys struct {
	Data  map[CustomInt]string
	Other map[struct{ X int }]string
}

// 60. Generic with pointer slice
type ResultWithPointerAndList struct {
	Data GenericResult[*UserAccount]         `json:"data"`
	List GenericResult[[]*UserProfileDetail] `json:"list"`
}

// 61. Generic with map
type ResultWithMap struct {
	Mapping GenericResult[map[string]*UserProfileDetail] `json:"mapping"`
}

// 62. Outer generic
type NestedResultGeneric struct {
	Nested GenericResult[GenericResult[*UserAccount]] `json:"nested"`
}

// 63. Complex generic chain
type MapSliceGenericResultChain struct {
	ComplexField map[string][]GenericResult[*UserAccount]         `json:"complex_field"`
	DeepNested   []map[string]GenericResult[[]*UserProfileDetail] `json:"deep_nested"`
}

// 64. Alias map result
type AliasMapResultType = GenericResult[AliasMapType]

// 65. Alias map result struct
type ResultWithAliasMap struct {
	AliasResult AliasMapResultType `json:"alias_result"`
}

// 66. ErrorTest tests conversion of a simple struct with an error field.
type ErrorTest struct {
	Err error `json:"err"`
}

// 67. Response is a generic wrapper with a payload and an error.
type Response[T any] struct {
	Payload T     `json:"payload"`
	Err     error `json:"err"`
}

// ComplexErrorTest combines simple and generic structs containing error fields.
type ComplexErrorTest struct {
	SingleError ErrorTest        `json:"single_error"`
	GenericResp Response[string] `json:"generic_resp"`
	GenericErr  Response[error]  `json:"generic_err"`
}
