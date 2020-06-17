
<a name="v1.13.0"></a>
## [v1.13.0](https://github.com/swipe-io/swipe/compare/v1.12.1...v1.13.0) (2020-06-17)

### Features

* Transport moved to external dependency in JS client generation
* For REST transport, added default error generation with implementation of the StatusCoder interface


<a name="v1.12.1"></a>
## [v1.12.1](https://github.com/swipe-io/swipe/compare/v1.12.0...v1.12.1) (2020-06-15)

### Bug Fixes

* added ints types
* generate config for []string type


<a name="v1.12.0"></a>
## [v1.12.0](https://github.com/swipe-io/swipe/compare/v1.11.4...v1.12.0) (2020-06-15)

### Bug Fixes

* openapi generation adds Uint, Uint8, Uint16, Uint32, Uint64 types
* added Interface and Map type for make openapi schema
* added check for named results greater than 1
* the absence of a comma in the named result of the method in the generation of endpoints

### Features

* added js client generation for jsonrpc


<a name="v1.11.4"></a>
## [v1.11.4](https://github.com/swipe-io/swipe/compare/v1.11.2...v1.11.4) (2020-06-10)

### Bug Fixes

* an example was added to the openapi documentation generation for the []byte type
* added for openapi encoding/json.RawMessage type, interpretate as object


<a name="v1.11.2"></a>
## [v1.11.2](https://github.com/swipe-io/swipe/compare/v1.11.0...v1.11.2) (2020-06-10)

### Bug Fixes

* default return name field for makeLogParam, and added Chan type
* generate endpoint for non named results


<a name="v1.11.0"></a>
## [v1.11.0](https://github.com/swipe-io/swipe/compare/v1.10.0...v1.11.0) (2020-06-10)

### Bug Fixes

* generate precision is -1 for FormatFloat
* WrapResponse option type changed to MethodOption
* MethodDefaultOptions not work if the MethodOptions option has not been set
* generating a value for the request/response if there are no values/result for the method

### BREAKING CHANGE


with version <= 1.10.0


<a name="v1.10.0"></a>
## [v1.10.0](https://github.com/swipe-io/swipe/compare/v1.9.0...v1.10.0) (2020-06-08)

### Bug Fixes

* FormatFloat incorrect fmt arg generate
* the base path in the client's request was erased
* not generate middlewareChain if server generate disabled
* not generate request typecast for rest client bug
* make query vars mapping using incorrect index

### Features

* renamed NotWrapBody option to WrapResponse, documentation added

### BREAKING CHANGE


NotWrapBody is not compatible with version <= 1.9.0


<a name="v1.9.0"></a>
## [v1.9.0](https://github.com/swipe-io/swipe/compare/v1.8.0...v1.9.0) (2020-06-03)

### Bug Fixes

* Moved OpenapiTags, OpenapiErrors from MethodOptions to Openapi

### BREAKING CHANGE


Changes for OpenapiTags, OpenapiErrors are not compatible with versions <v1.9.0


<a name="v1.8.0"></a>
## [v1.8.0](https://github.com/swipe-io/swipe/compare/v1.7.2...v1.8.0) (2020-06-03)

### Features

* removed copy code from swipe generation file

### BREAKING CHANGE


if you used the ability to use the code in the generation description file, then the update will not work


<a name="v1.7.2"></a>
## [v1.7.2](https://github.com/swipe-io/swipe/compare/v1.6.0...v1.7.2) (2020-06-02)

### Bug Fixes

* remove unused constants
* added to generate file swipe version

### Features

* added MethodDefaultOptions for sets default methods options


<a name="v1.6.0"></a>
## [v1.6.0](https://github.com/swipe-io/swipe/compare/v1.3.0...v1.6.0) (2020-06-02)

### Bug Fixes

* doc generation with NotWrapBody option

### Features

* added OpenapiTags option for set tags method on openapi doc generation
* added OpenapiErrors for maping errors to method on openapi docs
* added notWrapBody for JSON RPC generation


<a name="v1.3.0"></a>
## [v1.3.0](https://github.com/swipe-io/swipe/compare/v1.2.2...v1.3.0) (2020-05-31)

### Features

* update gokit jsonrpc transport to v1.10.0, added batch requests


<a name="v1.2.2"></a>
## [v1.2.2](https://github.com/swipe-io/swipe/compare/v1.1.5...v1.2.2) (2020-05-29)

### Bug Fixes

* not worked options ServerEncodeResponseFunc, ServerDecodeRequestFunc, ClientEncodeRequestFunc, ClientDecodeResponseFunc
* generate return error for endpoint and encode/decode func

### Features

* added generate set generic endpoint middlewares option


<a name="v1.1.5"></a>
## [v1.1.5](https://github.com/swipe-io/swipe/compare/v1.1.4...v1.1.5) (2020-05-27)

### Bug Fixes

* in transport generation, error handling in endpoint


<a name="v1.1.4"></a>
## [v1.1.4](https://github.com/swipe-io/swipe/compare/v1.1.3...v1.1.4) (2020-05-22)

### Bug Fixes

* invalid type generation in the example field. added generation of standard JSON RPC errors


<a name="v1.1.3"></a>
## [v1.1.3](https://github.com/swipe-io/swipe/compare/v1.1.2...v1.1.3) (2020-05-21)

### Bug Fixes

* search for errors to generate error mapping codes


<a name="v1.1.2"></a>
## [v1.1.2](https://github.com/swipe-io/swipe/compare/v1.1.1...v1.1.2) (2020-05-21)

### Bug Fixes

* generate server middileware option funcs


<a name="v1.1.1"></a>
## [v1.1.1](https://github.com/swipe-io/swipe/compare/v1.1.0...v1.1.1) (2020-05-21)


<a name="v1.1.0"></a>
## [v1.1.0](https://github.com/swipe-io/swipe/compare/v1.0.5...v1.1.0) (2020-05-21)

### Features

* implement generate endpoint middleware option


<a name="v1.0.5"></a>
## [v1.0.5](https://github.com/swipe-io/swipe/compare/v1.0.4...v1.0.5) (2020-05-21)


<a name="v1.0.4"></a>
## [v1.0.4](https://github.com/swipe-io/swipe/compare/v1.0.3...v1.0.4) (2020-05-21)

### Bug Fixes

* openapi doc generated bugs, replace options OpenapiVersion, OpenapiTitle, OpenapiDescription to OpenapiInfo(title, description, version)


<a name="v1.0.3"></a>
## [v1.0.3](https://github.com/swipe-io/swipe/compare/v1.0.2...v1.0.3) (2020-05-20)

### Bug Fixes

* not ignore no error types


<a name="v1.0.2"></a>
## [v1.0.2](https://github.com/swipe-io/swipe/compare/v1.0.1...v1.0.2) (2020-05-20)


<a name="v1.0.1"></a>
## [v1.0.1](https://github.com/swipe-io/swipe/compare/v1.0.0...v1.0.1) (2020-05-20)


<a name="v1.0.0"></a>
## v1.0.0 (2020-05-19)

