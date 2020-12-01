<a name="unreleased"></a>
## [Unreleased]


<a name="v2.0.0-beta6"></a>
## [v2.0.0-beta6] - 2020-11-27
### Bug Fixes
- crash generate if method not return error
- do not use body for rest server delete method


<a name="v2.0.0-beta5"></a>
## [v2.0.0-beta5] - 2020-11-20
### Bug Fixes
- rest client generating fmt import when not in use


<a name="v2.0.0-beta4"></a>
## [v2.0.0-beta4] - 2020-11-20
### Bug Fixes
- added for REST server generate use body


<a name="v2.0.0-beta3"></a>
## [v2.0.0-beta3] - 2020-11-20

<a name="v2.0.0-beta2"></a>
## [v2.0.0-beta2] - 2020-11-20
### Bug Fixes
- json rcp server generating fmt import when not in use


<a name="v2.0.0-beta1"></a>
## [v2.0.0-beta1] - 2020-11-20

<a name="2.0.0-alpha.23"></a>
## [2.0.0-alpha.23] - 2020-11-20

<a name="2.0.0-alpha.22"></a>
## [2.0.0-alpha.22] - 2020-11-20

<a name="v2.0.0-alpha.22"></a>
## [v2.0.0-alpha.22] - 2020-11-20
### Bug Fixes
- invalid class name for JavaScript JSON RPC client
- logging for UUID type
- generating fmt imports when not in use


<a name="v2.0.0-alpha.21"></a>
## [v2.0.0-alpha.21] - 2020-11-09

<a name="v2.0.0-alpha.20"></a>
## [v2.0.0-alpha.20] - 2020-11-09
### Bug Fixes
- a REST client was generated with request marshaling to the body without method parameters


<a name="v2.0.0-alpha.19"></a>
## [v2.0.0-alpha.19] - 2020-11-06
### Features
- added DefaultErrorEncoder option for set server error encoder


<a name="v2.0.0-alpha.18"></a>
## [v2.0.0-alpha.18] - 2020-11-03
### Bug Fixes
- binding comments to method and structure fields


<a name="v2.0.0-alpha.17"></a>
## [v2.0.0-alpha.17] - 2020-11-03
### Bug Fixes
- generate REST client default HTTP method use POST instead GET
- generate openapi JSON RPC method prefix
- disable generate convert codes to errors for REST

### Features
- added data field in errorWrapper if error implement interface{ ErrorData() interface{} }


<a name="v2.0.0-alpha.16"></a>
## [v2.0.0-alpha.16] - 2020-11-02
### Bug Fixes
- generating method errors for REST in openapi documentation
- generating the name of an environment variable containing a number, working with required, changing settings tags
- incorrect use of unwrap, added interface {LogError () error} instead
- REST error message is not wrapped in JSON object
- generate rest go client, excess newline

### Features
- added fix-comment command
- remove version command

### BREAKING CHANGE

settings tags, all settings are now in the env tag, such as required, use_zero (do not check empty value when checking required), desc: <value>, use_flag


<a name="v2.0.0-alpha.15"></a>
## [v2.0.0-alpha.15] - 2020-10-28
### Bug Fixes
- order export class for js jsonrpc client


<a name="v2.0.0-alpha.14"></a>
## [v2.0.0-alpha.14] - 2020-10-28

<a name="v2.0.0-alpha.13"></a>
## [v2.0.0-alpha.13] - 2020-10-28
### Bug Fixes
- rest server prefix generate without slash
- generate not correct class name for js jsonrpc client
- fixed check error in logging middleware

### Features
- added generate .gitattributes file for add ignore git dif

### Pull Requests
- Merge pull request [#12](https://github.com/swipe-io/swipe/issues/12) from sah4ez/v2


<a name="v2.0.0-alpha.12"></a>
## [v2.0.0-alpha.12] - 2020-10-27
### Bug Fixes
- openapi embedded structs


<a name="v2.0.0-alpha.11"></a>
## [v2.0.0-alpha.11] - 2020-10-27
### Bug Fixes
- generate logging for one named result param
- generate description for openapi methods
- fixed filling opts for sub-itnerfaces

### Pull Requests
- Merge pull request [#10](https://github.com/swipe-io/swipe/issues/10) from sah4ez/v2


<a name="v2.0.0-alpha.10"></a>
## [v2.0.0-alpha.10] - 2020-10-26
### Bug Fixes
- prettier format
- invalid help info for `swipe gen -h`
- failed generation specifying the file


<a name="v2.0.0-alpha.9"></a>
## [v2.0.0-alpha.9] - 2020-10-26
### Bug Fixes
- when generating a "panic" error if the project is empty


<a name="v2.0.0-alpha.8"></a>
## [v2.0.0-alpha.8] - 2020-10-26
### Bug Fixes
- when generating a "panic" error if the project is empty


<a name="v2.0.0-alpha.7"></a>
## [v2.0.0-alpha.7] - 2020-10-23
### Bug Fixes
- change func names generate and REST path values
- generate format for REST query string and headers and refactoring WriteConvertType


<a name="v2.0.0-alpha.6"></a>
## [v2.0.0-alpha.6] - 2020-10-21
### Features
- added multiple service interfaces, added time.Time type for REST query convert generate, added time.Duration for config generate, some not noticeable improvements.
- added output format error message

### BREAKING CHANGE

the format of the settings description has been changed, see the file github.com/swipe-io/swipe/swipe.go for details.


<a name="v2.0.0-alpha.5"></a>
## [v2.0.0-alpha.5] - 2020-10-06
### Bug Fixes
- Generate js type for alias and type definitions.

### Features
- Logging and Instrumenting moved to method options, added LoggingParams option allowing to enable or disable field logging.

### BREAKING CHANGE

Logging and Instrumenting cannot be used as ServiceOption.


<a name="v2.0.0-alpha.4"></a>
## [v2.0.0-alpha.4] - 2020-09-28
### Bug Fixes
- Incorrect project generation for files with $struct in the template file name.

### Features
- Added the ability to generate circular structures for openapi and JavaScript JSON RPC client.
- Rename command crud-service to gen-tpl.


<a name="v2.0.0-alpha.3"></a>
## [v2.0.0-alpha.3] - 2020-09-16
### Bug Fixes
- Not use json tag for struct in openapi generation.


<a name="v2.0.0-alpha.2"></a>
## [v2.0.0-alpha.2] - 2020-09-16
### Bug Fixes
- Show notifications when generated.


<a name="v2.0.0-alpha.1"></a>
## [v2.0.0-alpha.1] - 2020-09-15
### Features
- Some internal changes, performance improvements.

### BREAKING CHANGE

The swipe functions are no longer available in the github.com/swipe-io/swipe/pkg/swipe package, they are now located at github.com/swipe-io/swipe.


<a name="v1.26.7"></a>
## [v1.26.7] - 2020-09-09

<a name="vv2.0.0-alpha.15"></a>
## [vv2.0.0-alpha.15] - 2020-09-09

<a name="vv2.0.0-alpha.16"></a>
## [vv2.0.0-alpha.16] - 2020-09-09
### Features
- Added Path property to EndpointFactory for concatenated to server URL.


<a name="v1.26.6"></a>
## [v1.26.6] - 2020-09-08
### Bug Fixes
- Check url scheme for go cloent generation.

### Pull Requests
- Merge pull request [#2](https://github.com/swipe-io/swipe/issues/2) from djeckson/fix/check-url-scheme


<a name="v1.26.5"></a>
## [v1.26.5] - 2020-09-07
### Bug Fixes
- Generate check for the "http" prefix in JSON RPC/Rest client.


<a name="v1.26.4"></a>
## [v1.26.4] - 2020-09-07
### Features
- Added in Rest client splits a network address ([host]:port and more) and URL schema if not exists.
- Added in JSON RPC client splits a network address ([host]:port and more) and URL schema if not exists.


<a name="v1.26.3"></a>
## [v1.26.3] - 2020-09-04
### Bug Fixes
- Work with pointer errors.


<a name="v1.26.2"></a>
## [v1.26.2] - 2020-09-04
### Bug Fixes
- Return result generate.

### Features
- Added data parameter to ErrorDecode for JSON RPC.


<a name="v1.26.1"></a>
## [v1.26.1] - 2020-09-03
### Bug Fixes
- Generation of the returned result if it is not in the method.
- Added _gen for the name of the generated file.
- Deleting files with the "_gen" pattern before create new generated files.


<a name="v1.26.0"></a>
## [v1.26.0] - 2020-08-28
### Features
- Update l-vitaly/go-kit to v1.12.0.

### BREAKING CHANGE

Now all requests are not asynchronous, to enable an asynchronous request you need to pass in the header: "X-Async: on"


<a name="v1.25.12"></a>
## [v1.25.12] - 2020-08-28
### Bug Fixes
- Incorrect named result in JSDoc JSON RPC client.
- Use "key" for map key name.


<a name="v1.25.11"></a>
## [v1.25.11] - 2020-08-27
### Bug Fixes
- Returning deleted code.


<a name="v1.25.10"></a>
## [v1.25.10] - 2020-08-27
### Bug Fixes
- Incorrect generate with named results.


<a name="v1.25.9"></a>
## [v1.25.9] - 2020-08-16
### Bug Fixes
- Use pg.NullTime type instead of sql.NullTime when generating CRUD service.


<a name="v1.25.8"></a>
## [v1.25.8] - 2020-08-15
### Features
- Updated the generation of JavaScript JSON RPC client documentation.


<a name="v1.25.7"></a>
## [v1.25.7] - 2020-08-15
### Features
- Added use of goimports to format Golang code if installed.


<a name="v1.25.6"></a>
## [v1.25.6] - 2020-08-15
### Features
- Added generation of examples of using the JS RPC client in the markdown documentation.


<a name="v1.25.5"></a>
## [v1.25.5] - 2020-08-15
### Bug Fixes
- Added checking for golang standard named types when generating JSON RPC documentation.


<a name="v1.25.4"></a>
## [v1.25.4] - 2020-08-15
### Bug Fixes
- Added check for nil Object when loading AST tree.

### Features
- Added output of contents of go file in which there is an error.


<a name="v1.25.3"></a>
## [v1.25.3] - 2020-08-13
### Bug Fixes
- Config flag generate.


<a name="v1.25.2"></a>
## [v1.25.2] - 2020-08-11
### Features
- Added support for embedded structures for generating documentation.


<a name="v1.25.1"></a>
## [v1.25.1] - 2020-08-11
### Bug Fixes
- Added sorting for errors in the ErrorDecode method.
- Added check type when searching enums.


<a name="v1.25.0"></a>
## [v1.25.0] - 2020-08-10
### Features
- Added config markdown generate option ConfigMarkdownDoc.
- Added Name option for override service name prefix.


<a name="v1.24.7"></a>
## [v1.24.7] - 2020-08-08
### Bug Fixes
- Added check if template path not exists in crud-service command.

### Features
- Removed the -entities flag from the crud-service command and added the -config flag, moved the entity loader to the config.


<a name="v1.24.6"></a>
## [v1.24.6] - 2020-08-06
### Bug Fixes
- Generation of service transport if there are no interface methods.
- To generate the CRUD service, the argument for setting the path to the entity description file has been changed to the -entities flag.

### Features
- For default readme template added git versions section.


<a name="v1.24.5"></a>
## [v1.24.5] - 2020-08-04
### Bug Fixes
- Generate section Members ans Enums if exists items.
- Added annotation for encoding/json.RawMessage in markdown JS client documentation.
- Added annotation for encoding/json.RawMessage in JS client.
- Invalid path definition for JSON RPC client documentation.
- Incorrect generate enum name for markdown docs.

### Features
- Added command for generate CRUD service structure.


<a name="v1.24.3"></a>
## [v1.24.3] - 2020-07-28
### Bug Fixes
- Change enums JSON RPC client generate format.


<a name="v1.24.2"></a>
## [v1.24.2] - 2020-07-28
### Bug Fixes
- Added generate enum constants for JSON RPC client.


<a name="v1.24.1"></a>
## [v1.24.1] - 2020-07-28

<a name="v1.24.4"></a>
## [v1.24.4] - 2020-07-28
### Bug Fixes
- Incorrect generate enum name for markdown docs.
- Change enums JSON RPC client generate format.
- Added generate enum constants for JSON RPC client.

### Features
- Added readme markdown and JSON RPC client markdown documentation.


<a name="v1.24.0"></a>
## [v1.24.0] - 2020-07-21
### Features
- Implement new graph type for graph types and optimize performance.


<a name="v1.23.0"></a>
## [v1.23.0] - 2020-07-20
### Bug Fixes
- Encode/decode function was not is generated correctly.

### Features
- Added simple gateway gokit helpers generator.
- Changed the prefix in the generated code, instead of <packageName><serviceName>, <projectName><serviceName>is used

### BREAKING CHANGE

Generated exported functions will have different names.


<a name="v1.22.4"></a>
## [v1.22.4] - 2020-07-07
### Bug Fixes
- Looping recursion when looking for error types.


<a name="v1.22.3"></a>
## [v1.22.3] - 2020-07-06
### Bug Fixes
- The order of the parameters specified in the Path option was not respected.


<a name="v1.22.2"></a>
## [v1.22.2] - 2020-07-06
### Bug Fixes
- JSON RPC JS client change var to const.
- JSON RPC JS client change var to let.
- JSON RPC JS client added check hasOwnProperty.


<a name="v1.22.1"></a>
## [v1.22.1] - 2020-07-06
### Bug Fixes
- JSON RPC JS client generate catch block.


<a name="v1.22.0"></a>
## [v1.22.0] - 2020-07-06
### Features
- Added automatic generation error mapping for Gokit service and Openapi docs.

### BREAKING CHANGE

Remove OpenapiErrors openapi option.


<a name="v1.21.0"></a>
## [v1.21.0] - 2020-06-29
### Bug Fixes
- ConfigEnv imports generation

### Features
- Added default generation of requestCount and requestLatency for Instrumenting.

### BREAKING CHANGE

The Namespace and Subsystem options have been removed, now the values are passed as parameters for Instrumenting.


<a name="v1.20.1"></a>
## [v1.20.1] - 2020-06-29
### Bug Fixes
- Remove Assembly option and generation

### Features
- The generated file name for the JSON RPC JavaScript client has been changed to jsonrpc_client_gen.js.
- Added verification of the correspondence of the version of swipe cli and swipe package.
- Added factory generation for gokit.Endpoint.


<a name="v1.14.0"></a>
## [v1.14.0] - 2020-06-24
### Features
- Added comment generation from the service interface for openapi documentation and JS client


<a name="v1.13.4"></a>
## [v1.13.4] - 2020-06-23
### Bug Fixes
- An extra wrapper was added for the returned parameter


<a name="v1.13.3"></a>
## [v1.13.3] - 2020-06-23
### Bug Fixes
- Added map type for Openapi and JS client generate


<a name="v1.13.2"></a>
## [v1.13.2] - 2020-06-23
### Bug Fixes
- generate JS client for wrapped data
- openapi and JS client generate pointer type


<a name="v1.13.1"></a>
## [v1.13.1] - 2020-06-23
### Bug Fixes
- Type cast response for rest/jsonrpc
- Added ignore unexported errors


<a name="v1.13.0"></a>
## [v1.13.0] - 2020-06-17
### Features
- Transport moved to external dependency in JS client generation
- For REST transport, added default error generation with implementation of the StatusCoder interface


<a name="v1.12.1"></a>
## [v1.12.1] - 2020-06-15
### Bug Fixes
- added ints types
- generate config for []string type


<a name="v1.12.0"></a>
## [v1.12.0] - 2020-06-15
### Bug Fixes
- openapi generation adds Uint, Uint8, Uint16, Uint32, Uint64 types
- added Interface and Map type for make openapi schema
- added check for named results greater than 1
- the absence of a comma in the named result of the method in the generation of endpoints

### Features
- added js client generation for jsonrpc


<a name="v1.11.4"></a>
## [v1.11.4] - 2020-06-10
### Bug Fixes
- an example was added to the openapi documentation generation for the []byte type
- added for openapi encoding/json.RawMessage type, interpretate as object


<a name="v1.11.2"></a>
## [v1.11.2] - 2020-06-10
### Bug Fixes
- default return name field for makeLogParam, and added Chan type
- generate endpoint for non named results


<a name="v1.11.0"></a>
## [v1.11.0] - 2020-06-10
### Bug Fixes
- generate precision is -1 for FormatFloat
- WrapResponse option type changed to MethodOption
- MethodDefaultOptions not work if the MethodOptions option has not been set
- generating a value for the request/response if there are no values/result for the method

### BREAKING CHANGE

with version <= 1.10.0


<a name="v1.10.0"></a>
## [v1.10.0] - 2020-06-08
### Bug Fixes
- FormatFloat incorrect fmt arg generate
- the base path in the client's request was erased
- not generate middlewareChain if server generate disabled
- not generate request typecast for rest client bug
- make query vars mapping using incorrect index

### Features
- renamed NotWrapBody option to WrapResponse, documentation added

### BREAKING CHANGE

NotWrapBody is not compatible with version <= 1.9.0


<a name="v1.9.0"></a>
## [v1.9.0] - 2020-06-03
### Bug Fixes
- Moved OpenapiTags, OpenapiErrors from MethodOptions to Openapi

### BREAKING CHANGE

Changes for OpenapiTags, OpenapiErrors are not compatible with versions <v1.9.0


<a name="v1.8.0"></a>
## [v1.8.0] - 2020-06-03
### Features
- removed copy code from swipe generation file

### BREAKING CHANGE

if you used the ability to use the code in the generation description file, then the update will not work


<a name="v1.7.2"></a>
## [v1.7.2] - 2020-06-02
### Bug Fixes
- remove unused constants
- added to generate file swipe version

### Features
- added MethodDefaultOptions for sets default methods options


<a name="v1.6.0"></a>
## [v1.6.0] - 2020-06-02
### Bug Fixes
- doc generation with NotWrapBody option

### Features
- added OpenapiTags option for set tags method on openapi doc generation
- added OpenapiErrors for maping errors to method on openapi docs
- added notWrapBody for JSON RPC generation


<a name="v1.3.0"></a>
## [v1.3.0] - 2020-05-31
### Features
- update gokit jsonrpc transport to v1.10.0, added batch requests


<a name="v1.2.2"></a>
## [v1.2.2] - 2020-05-29
### Bug Fixes
- not worked options ServerEncodeResponseFunc, ServerDecodeRequestFunc, ClientEncodeRequestFunc, ClientDecodeResponseFunc
- generate return error for endpoint and encode/decode func

### Features
- added generate set generic endpoint middlewares option


<a name="v1.1.5"></a>
## [v1.1.5] - 2020-05-27
### Bug Fixes
- in transport generation, error handling in endpoint


<a name="v1.1.4"></a>
## [v1.1.4] - 2020-05-22
### Bug Fixes
- invalid type generation in the example field. added generation of standard JSON RPC errors


<a name="v1.1.3"></a>
## [v1.1.3] - 2020-05-21
### Bug Fixes
- search for errors to generate error mapping codes


<a name="v1.1.2"></a>
## [v1.1.2] - 2020-05-21
### Bug Fixes
- generate server middileware option funcs


<a name="v1.1.1"></a>
## [v1.1.1] - 2020-05-21

<a name="v1.1.0"></a>
## [v1.1.0] - 2020-05-21
### Features
- implement generate endpoint middleware option


<a name="v1.0.5"></a>
## [v1.0.5] - 2020-05-21

<a name="v1.0.4"></a>
## [v1.0.4] - 2020-05-21
### Bug Fixes
- openapi doc generated bugs, replace options OpenapiVersion, OpenapiTitle, OpenapiDescription to OpenapiInfo(title, description, version)


<a name="v1.0.3"></a>
## [v1.0.3] - 2020-05-20
### Bug Fixes
- not ignore no error types


<a name="v1.0.2"></a>
## [v1.0.2] - 2020-05-20

<a name="v1.0.1"></a>
## [v1.0.1] - 2020-05-20

<a name="v1.0.0"></a>
## v1.0.0 - 2020-05-19

[Unreleased]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta6...HEAD
[v2.0.0-beta6]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta5...v2.0.0-beta6
[v2.0.0-beta5]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta4...v2.0.0-beta5
[v2.0.0-beta4]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta3...v2.0.0-beta4
[v2.0.0-beta3]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta2...v2.0.0-beta3
[v2.0.0-beta2]: https://github.com/swipe-io/swipe/compare/v2.0.0-beta1...v2.0.0-beta2
[v2.0.0-beta1]: https://github.com/swipe-io/swipe/compare/2.0.0-alpha.23...v2.0.0-beta1
[2.0.0-alpha.23]: https://github.com/swipe-io/swipe/compare/2.0.0-alpha.22...2.0.0-alpha.23
[2.0.0-alpha.22]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.22...2.0.0-alpha.22
[v2.0.0-alpha.22]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.21...v2.0.0-alpha.22
[v2.0.0-alpha.21]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.20...v2.0.0-alpha.21
[v2.0.0-alpha.20]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.19...v2.0.0-alpha.20
[v2.0.0-alpha.19]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.18...v2.0.0-alpha.19
[v2.0.0-alpha.18]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.17...v2.0.0-alpha.18
[v2.0.0-alpha.17]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.16...v2.0.0-alpha.17
[v2.0.0-alpha.16]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.15...v2.0.0-alpha.16
[v2.0.0-alpha.15]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.14...v2.0.0-alpha.15
[v2.0.0-alpha.14]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.13...v2.0.0-alpha.14
[v2.0.0-alpha.13]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.12...v2.0.0-alpha.13
[v2.0.0-alpha.12]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.11...v2.0.0-alpha.12
[v2.0.0-alpha.11]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.10...v2.0.0-alpha.11
[v2.0.0-alpha.10]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.9...v2.0.0-alpha.10
[v2.0.0-alpha.9]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.8...v2.0.0-alpha.9
[v2.0.0-alpha.8]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.7...v2.0.0-alpha.8
[v2.0.0-alpha.7]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.6...v2.0.0-alpha.7
[v2.0.0-alpha.6]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.5...v2.0.0-alpha.6
[v2.0.0-alpha.5]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.4...v2.0.0-alpha.5
[v2.0.0-alpha.4]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.3...v2.0.0-alpha.4
[v2.0.0-alpha.3]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.2...v2.0.0-alpha.3
[v2.0.0-alpha.2]: https://github.com/swipe-io/swipe/compare/v2.0.0-alpha.1...v2.0.0-alpha.2
[v2.0.0-alpha.1]: https://github.com/swipe-io/swipe/compare/v1.26.7...v2.0.0-alpha.1
[v1.26.7]: https://github.com/swipe-io/swipe/compare/vv2.0.0-alpha.15...v1.26.7
[vv2.0.0-alpha.15]: https://github.com/swipe-io/swipe/compare/vv2.0.0-alpha.16...vv2.0.0-alpha.15
[vv2.0.0-alpha.16]: https://github.com/swipe-io/swipe/compare/v1.26.6...vv2.0.0-alpha.16
[v1.26.6]: https://github.com/swipe-io/swipe/compare/v1.26.5...v1.26.6
[v1.26.5]: https://github.com/swipe-io/swipe/compare/v1.26.4...v1.26.5
[v1.26.4]: https://github.com/swipe-io/swipe/compare/v1.26.3...v1.26.4
[v1.26.3]: https://github.com/swipe-io/swipe/compare/v1.26.2...v1.26.3
[v1.26.2]: https://github.com/swipe-io/swipe/compare/v1.26.1...v1.26.2
[v1.26.1]: https://github.com/swipe-io/swipe/compare/v1.26.0...v1.26.1
[v1.26.0]: https://github.com/swipe-io/swipe/compare/v1.25.12...v1.26.0
[v1.25.12]: https://github.com/swipe-io/swipe/compare/v1.25.11...v1.25.12
[v1.25.11]: https://github.com/swipe-io/swipe/compare/v1.25.10...v1.25.11
[v1.25.10]: https://github.com/swipe-io/swipe/compare/v1.25.9...v1.25.10
[v1.25.9]: https://github.com/swipe-io/swipe/compare/v1.25.8...v1.25.9
[v1.25.8]: https://github.com/swipe-io/swipe/compare/v1.25.7...v1.25.8
[v1.25.7]: https://github.com/swipe-io/swipe/compare/v1.25.6...v1.25.7
[v1.25.6]: https://github.com/swipe-io/swipe/compare/v1.25.5...v1.25.6
[v1.25.5]: https://github.com/swipe-io/swipe/compare/v1.25.4...v1.25.5
[v1.25.4]: https://github.com/swipe-io/swipe/compare/v1.25.3...v1.25.4
[v1.25.3]: https://github.com/swipe-io/swipe/compare/v1.25.2...v1.25.3
[v1.25.2]: https://github.com/swipe-io/swipe/compare/v1.25.1...v1.25.2
[v1.25.1]: https://github.com/swipe-io/swipe/compare/v1.25.0...v1.25.1
[v1.25.0]: https://github.com/swipe-io/swipe/compare/v1.24.7...v1.25.0
[v1.24.7]: https://github.com/swipe-io/swipe/compare/v1.24.6...v1.24.7
[v1.24.6]: https://github.com/swipe-io/swipe/compare/v1.24.5...v1.24.6
[v1.24.5]: https://github.com/swipe-io/swipe/compare/v1.24.3...v1.24.5
[v1.24.3]: https://github.com/swipe-io/swipe/compare/v1.24.2...v1.24.3
[v1.24.2]: https://github.com/swipe-io/swipe/compare/v1.24.1...v1.24.2
[v1.24.1]: https://github.com/swipe-io/swipe/compare/v1.24.4...v1.24.1
[v1.24.4]: https://github.com/swipe-io/swipe/compare/v1.24.0...v1.24.4
[v1.24.0]: https://github.com/swipe-io/swipe/compare/v1.23.0...v1.24.0
[v1.23.0]: https://github.com/swipe-io/swipe/compare/v1.22.4...v1.23.0
[v1.22.4]: https://github.com/swipe-io/swipe/compare/v1.22.3...v1.22.4
[v1.22.3]: https://github.com/swipe-io/swipe/compare/v1.22.2...v1.22.3
[v1.22.2]: https://github.com/swipe-io/swipe/compare/v1.22.1...v1.22.2
[v1.22.1]: https://github.com/swipe-io/swipe/compare/v1.22.0...v1.22.1
[v1.22.0]: https://github.com/swipe-io/swipe/compare/v1.21.0...v1.22.0
[v1.21.0]: https://github.com/swipe-io/swipe/compare/v1.20.1...v1.21.0
[v1.20.1]: https://github.com/swipe-io/swipe/compare/v1.14.0...v1.20.1
[v1.14.0]: https://github.com/swipe-io/swipe/compare/v1.13.4...v1.14.0
[v1.13.4]: https://github.com/swipe-io/swipe/compare/v1.13.3...v1.13.4
[v1.13.3]: https://github.com/swipe-io/swipe/compare/v1.13.2...v1.13.3
[v1.13.2]: https://github.com/swipe-io/swipe/compare/v1.13.1...v1.13.2
[v1.13.1]: https://github.com/swipe-io/swipe/compare/v1.13.0...v1.13.1
[v1.13.0]: https://github.com/swipe-io/swipe/compare/v1.12.1...v1.13.0
[v1.12.1]: https://github.com/swipe-io/swipe/compare/v1.12.0...v1.12.1
[v1.12.0]: https://github.com/swipe-io/swipe/compare/v1.11.4...v1.12.0
[v1.11.4]: https://github.com/swipe-io/swipe/compare/v1.11.2...v1.11.4
[v1.11.2]: https://github.com/swipe-io/swipe/compare/v1.11.0...v1.11.2
[v1.11.0]: https://github.com/swipe-io/swipe/compare/v1.10.0...v1.11.0
[v1.10.0]: https://github.com/swipe-io/swipe/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/swipe-io/swipe/compare/v1.8.0...v1.9.0
[v1.8.0]: https://github.com/swipe-io/swipe/compare/v1.7.2...v1.8.0
[v1.7.2]: https://github.com/swipe-io/swipe/compare/v1.6.0...v1.7.2
[v1.6.0]: https://github.com/swipe-io/swipe/compare/v1.3.0...v1.6.0
[v1.3.0]: https://github.com/swipe-io/swipe/compare/v1.2.2...v1.3.0
[v1.2.2]: https://github.com/swipe-io/swipe/compare/v1.1.5...v1.2.2
[v1.1.5]: https://github.com/swipe-io/swipe/compare/v1.1.4...v1.1.5
[v1.1.4]: https://github.com/swipe-io/swipe/compare/v1.1.3...v1.1.4
[v1.1.3]: https://github.com/swipe-io/swipe/compare/v1.1.2...v1.1.3
[v1.1.2]: https://github.com/swipe-io/swipe/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/swipe-io/swipe/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/swipe-io/swipe/compare/v1.0.5...v1.1.0
[v1.0.5]: https://github.com/swipe-io/swipe/compare/v1.0.4...v1.0.5
[v1.0.4]: https://github.com/swipe-io/swipe/compare/v1.0.3...v1.0.4
[v1.0.3]: https://github.com/swipe-io/swipe/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/swipe-io/swipe/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/swipe-io/swipe/compare/v1.0.0...v1.0.1
