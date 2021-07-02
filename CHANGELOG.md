# Change Log

## [Unreleased](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/Unreleased) (2020-02-11)
[Full Changelog](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v1.2.0...Unreleased)

**Closed issues:**

- New release? [\#259](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/259)

**Merged pull requests:**

- Add unary client interceptor for validation [\#265](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/265) ([sashayakovtseva](https://github.com/sashayakovtseva))
- unlock mutex after incrementing reqCounter [\#264](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/264) ([isiddi4](https://github.com/isiddi4))
- feature\(tracing\): Add datadog support to injectOpenTracingIdsToTags [\#262](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/262) ([wdullaer](https://github.com/wdullaer))

## [v1.2.0](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v1.2.0) (2020-01-28)
[Full Changelog](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v1.1.0...v1.2.0)

**Implemented enhancements:**

- Logrus payload logging: Allow gogo marshaller [\#235](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/235)

**Fixed bugs:**

- Duplicate peer.address in zap logger output [\#120](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/120)

**Closed issues:**

- go-grpc-middleware/auth [\#253](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/253)
- Module declares its path as ... [\#252](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/252)
- Hide recovered panic message from client [\#251](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/251)
- How can i use custom error in my project ? [\#246](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/246)
- Privileges to allow more people to become maintainers. [\#231](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/231)
- Is there any way I can ignore some function use authentication middleware? [\#213](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/213)
- Any planned release soon? [\#210](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/210)
- raise error when install by dep [\#198](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/198)
- tracedClientStream.CloseSend\(\) doesn't finish the ClientSpan [\#196](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/196)
- make rpc\_ctxtags.SetInContext  public access [\#182](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/182)
- Duplicate peer.address in zap logger output ，bug  [\#146](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/146)

**Merged pull requests:**

- Renew deprecated code [\#257](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/257) ([khasanovbi](https://github.com/khasanovbi))
- Fix typo [\#256](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/256) ([khasanovbi](https://github.com/khasanovbi))
- Fix typo [\#255](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/255) ([jeffwidman](https://github.com/jeffwidman))
- Update slack link [\#254](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/254) ([johanbrandhorst](https://github.com/johanbrandhorst))
- fix BackoffExponential's godoc. [\#250](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/250) ([syokomiz](https://github.com/syokomiz))
- Fix panic in TagBasedRequestFieldExtractor when gogoproto.stdtime is used [\#249](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/249) ([nvx](https://github.com/nvx))
- recovery: add custom func on recovery example [\#247](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/247) ([Shikugawa](https://github.com/Shikugawa))
- Using strings.EualFold\(\) for efficient string comparison [\#243](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/243) ([vivekv96](https://github.com/vivekv96))
- Make SetInContext and NewTags to be public [\#242](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/242) ([mkorolyov](https://github.com/mkorolyov))
- Added JsonPbMarshalling interface [\#236](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/236) ([petervandenbroek](https://github.com/petervandenbroek))
- adding v2 grpc logger [\#234](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/234) ([domgreen](https://github.com/domgreen))
- Changeloglinks [\#233](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/233) ([domgreen](https://github.com/domgreen))
- CHANGELOG [\#232](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/232) ([domgreen](https://github.com/domgreen))
- Fix span finish on client stream close \(fix \#196\) [\#204](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/204) ([inn0kenty](https://github.com/inn0kenty))
- Fix function comments based on best practices from Effective Go [\#200](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/200) ([CodeLingoTeam](https://github.com/CodeLingoTeam))
- fixed duplicate custom&peer.address field [\#187](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/187) ([zxfishhack](https://github.com/zxfishhack))
- Fix typo in chain\_test.go [\#153](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/153) ([RJPercival](https://github.com/RJPercival))

## [v1.1.0](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v1.1.0) (2019-09-12)
[Full Changelog](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v1.0.0...v1.1.0)

**Implemented enhancements:**

- Rate Limit Interceptor  [\#134](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/134)
- Add if the trace was sampled to the ctxTags [\#184](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/184) ([domgreen](https://github.com/domgreen))

**Closed issues:**

- logrus module pulled in by test [\#228](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/228)
- grpclogger should implement LoggerV2 interface [\#219](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/219)
- Retry timeout will overwrite the origin timeout in ctx [\#216](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/216)
- retry not working when grpc-server close the connection [\#199](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/199)
- Question about package naming [\#197](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/197)
- Checksum Mismatch on go get [\#193](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/193)
- go modules support [\#185](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/185)
- FR: In auth package, add exclude path to Interceptor builder [\#176](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/176)
- Replace "golang.org/x/net/context" with "context" [\#175](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/175)
- FR: In retry package, pass context to custom retry handler [\#171](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/171)
- FR: In recovery package, pass context to custom recovery handler [\#168](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/168)
- Can I start auth proxy? [\#158](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/158)
- retry doesnt seems to work  [\#154](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/154)
- grpc\_retry:not work as a calloption in my situation [\#151](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/151)
- Downstream interceptors are not re-executed on retry [\#110](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/110)
- Can we get semver releases? [\#40](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/40)

**Merged pull requests:**

- Quick Fix of \#228 [\#229](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/229) ([domgreen](https://github.com/domgreen))
- remove last net/context reference will help fix \#175 [\#227](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/227) ([domgreen](https://github.com/domgreen))
- Move from dep to modules [\#226](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/226) ([domgreen](https://github.com/domgreen))
- Add go-kit logging middleware [\#223](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/223) ([adrien-f](https://github.com/adrien-f))
- gRPC Zap ReplaceGrpcLoggerV2 [\#221](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/221) ([kush-patel-hs](https://github.com/kush-patel-hs))
- Refactor interceptor chain functions [\#220](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/220) ([jun06t](https://github.com/jun06t))
- `middleare` -\> `middleware` [\#217](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/217) ([seanhagen](https://github.com/seanhagen))
- golang.org/x/net/context =\> context [\#201](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/201) ([houz42](https://github.com/houz42))
- Deprecation, typo fixes in logging [\#194](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/194) ([khasanovbi](https://github.com/khasanovbi))
- Update Deps  [\#191](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/191) ([PrasadG193](https://github.com/PrasadG193))
- Remove generation of docs [\#183](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/183) ([domgreen](https://github.com/domgreen))
- Ratelimit interceptor [\#181](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/181) ([ceshihao](https://github.com/ceshihao))
- Add context functions to retry and recover [\#172](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/172) ([johanbrandhorst](https://github.com/johanbrandhorst))
- Fix typo in log message [\#164](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/164) ([scripter-v](https://github.com/scripter-v))
- Added retry on server stream call [\#161](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/161) ([lonnblad](https://github.com/lonnblad))
- Add exponential backoff functions [\#152](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/152) ([polyfloyd](https://github.com/polyfloyd))
- Add jaeger support for ctxtags extraction [\#147](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/147) ([vporoshok](https://github.com/vporoshok))
- Adding CHANGELOG [\#143](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/143) ([domgreen](https://github.com/domgreen))

## [v1.0.0](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v1.0.0) (2018-05-02)
**Implemented enhancements:**

- Add ability to skip logging/change level of certain method calls \(e.g.  healthcheck\) [\#89](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/89)

**Fixed bugs:**

- grpc\_retry.WithMax\(\), how to use it properly?  [\#37](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/37)

**Closed issues:**

- grpc\_logging payloads embedded in the classic log output [\#135](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/135)
- Import path of logrus may cause build error [\#125](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/125)
- could you clean merged branch? [\#124](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/124)
- \[grpc\_logrus\] how to add error stack trace to log output [\#121](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/121)
- Precision Error on durationToMilliseconds [\#116](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/116)
- vendored package breaks public API [\#105](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/105)
- \[Question\] Is it possible to select which endpoints to apply MW to? [\#96](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/96)
- \[Question\] grpc\_ctxtags without interceptor [\#90](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/90)
- grpc\_retry: Why does the serverStreamingRetryingStream never empty its buffer? [\#88](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/88)
- Build fails due to metadata api changes [\#77](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/77)
- Performance suggestion for chain builders [\#74](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/74)
- Change Sirupsen imports to sirupsen [\#73](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/73)
- go get gives compilation errors [\#70](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/70)
- Ctx\_tags to populate both Request & Response Logging [\#66](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/66)
- Question, binary file streaming [\#65](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/65)
- open wiki pages. [\#63](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/63)
- \[zap\]UnaryClientInterceptor should can accept super struct [\#61](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/61)
- zap log needs log a struct. [\#60](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/60)
- Need a full example [\#59](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/59)
- support other language client? [\#49](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/49)
- grpc\_auth: AuthFuncOverride without dummy interceptor [\#41](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/41)
- Client Logging Interceptor defaults behave incorrectly [\#36](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/36)
- Race in metautils [\#21](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/21)
- Custom errors [\#17](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/17)
- Prometheus Metrics not provisionned with grpc\_auth [\#16](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/16)
- build a proxy to NATs with this [\#10](https://github.com/grpc-ecosystem/go-grpc-middleware/issues/10)

**Merged pull requests:**

- Delete .DS\_Store [\#141](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/141) ([xissy](https://github.com/xissy))
- adding a slack shield [\#140](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/140) ([domgreen](https://github.com/domgreen))
- Add check for docs [\#138](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/138) ([domgreen](https://github.com/domgreen))
- lint [\#137](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/137) ([chemidy](https://github.com/chemidy))
- Update DOC.md [\#136](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/136) ([liukgg](https://github.com/liukgg))
- fixing vet issues ... most relating to the Examples [\#132](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/132) ([domgreen](https://github.com/domgreen))
- Improve logging/\*/doc.go [\#130](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/130) ([matiasanaya](https://github.com/matiasanaya))
- ctxloggers for zap and logrus should live with logging [\#128](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/128) ([domgreen](https://github.com/domgreen))
- updating final log message to include status code [\#122](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/122) ([domgreen](https://github.com/domgreen))
- change peer.address logrus representaion. [\#118](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/118) ([altaastro](https://github.com/altaastro))
- Fix \#116 - sub millisecond duration should now be measured by default [\#117](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/117) ([domgreen](https://github.com/domgreen))
- If ctx has a deadline it will be added as a tag on Extract [\#115](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/115) ([domgreen](https://github.com/domgreen))
- adding grpc.start to logging interceptors [\#114](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/114) ([domgreen](https://github.com/domgreen))
- adding option to suppress interceptor logs [\#113](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/113) ([domgreen](https://github.com/domgreen))
- vet now passes and run from makefile [\#111](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/111) ([domgreen](https://github.com/domgreen))
- Remove vendor and gitignore [\#109](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/109) ([domgreen](https://github.com/domgreen))
- updating travis to use dep [\#108](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/108) ([domgreen](https://github.com/domgreen))
- ctxlogger split out from grpc\_logging [\#107](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/107) ([domgreen](https://github.com/domgreen))
- fixup.sh ignores vendor and have fixed the dependency on godoc2ghmd [\#106](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/106) ([domgreen](https://github.com/domgreen))
- update docs to advise to only Extract once [\#103](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/103) ([domgreen](https://github.com/domgreen))
- dep for management of packages [\#101](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/101) ([domgreen](https://github.com/domgreen))
- Support retrying gRPCs with chained interceptors [\#100](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/100) ([marcwilson-g](https://github.com/marcwilson-g))
- Spelling/grammar corrections for validator docs [\#99](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/99) ([thurt](https://github.com/thurt))
- typos [\#98](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/98) ([graphaelli](https://github.com/graphaelli))
- Add function AddFields that allows setting zap fields without using tags. [\#91](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/91) ([bakins](https://github.com/bakins))
- Actually use jsonpb for payload marshaling. [\#85](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/85) ([jgiles](https://github.com/jgiles))
- implement noopTags [\#80](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/80) ([flisky](https://github.com/flisky))
- Refactored middleware chain builders [\#78](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/78) ([alcore](https://github.com/alcore))
- Make zap grpclog not Fatal on a simple Print [\#76](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/76) ([htdvisser](https://github.com/htdvisser))
- grpc\_ctxtags: Tagging on client streamed requests [\#75](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/75) ([suprememoocow](https://github.com/suprememoocow))
- tracing/opentracing: add ClientSpanTagKey for client span tag extension. [\#72](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/72) ([bom-d-van](https://github.com/bom-d-van))
- s/WithPerCallTimeout/WithPerRetryTimeout/g [\#69](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/69) ([shouichi](https://github.com/shouichi))
- Update time.Ticker in retry.go [\#68](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/68) ([itizir](https://github.com/itizir))
- Use EmptyOption for retry interceptor [\#55](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/55) ([mwitkow](https://github.com/mwitkow))
- Renamed github.com/Sirupsen/logrus to github.com/sirupsen/logrus [\#54](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/54) ([suprememoocow](https://github.com/suprememoocow))
- Response payload is not dumped [\#50](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/50) ([kazegusuri](https://github.com/kazegusuri))
- Add example generation to docs. [\#46](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/46) ([mwitkow](https://github.com/mwitkow))
- better docs autogen using godoc2ghmd [\#45](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/45) ([mwitkow](https://github.com/mwitkow))
- Optimize return value for chained interceptors in length 0 and 1 cases [\#39](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/39) ([bufdev](https://github.com/bufdev))
- Remove metautils functions that call FromContext and NewContext [\#38](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/38) ([bufdev](https://github.com/bufdev))
- Request/Response Payload Loggers for Logrus and Zap [\#35](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/35) ([mwitkow](https://github.com/mwitkow))
- Fix/jitter [\#34](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/34) ([feliksik](https://github.com/feliksik))
- Add Client interceptors for logging: logrus and zap [\#33](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/33) ([mwitkow](https://github.com/mwitkow))
- Extract Logging Suites for easier testing in the future [\#32](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/32) ([mwitkow](https://github.com/mwitkow))
- Make duration log configurable [\#31](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/31) ([dackroyd](https://github.com/dackroyd))
- Add 'Recovery' middleware [\#30](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/30) ([dackroyd](https://github.com/dackroyd))
- logging: Change logging time to ms float [\#29](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/29) ([mwitkow](https://github.com/mwitkow))
- add nicemd and update all usage to inbound/outbound ctx [\#28](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/28) ([mwitkow](https://github.com/mwitkow))
- Feature/repo move [\#27](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/27) ([mwitkow](https://github.com/mwitkow))
- opentracing: add support for traceid in tags [\#26](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/26) ([mwitkow](https://github.com/mwitkow))
- Remove README.md [\#25](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/25) ([tomwilkie](https://github.com/tomwilkie))
- add opentracing interceptors for streams and unary [\#24](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/24) ([mwitkow](https://github.com/mwitkow))
- Breaking change to Logging: Add grpc\_ctxtags [\#23](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/23) ([mwitkow](https://github.com/mwitkow))
- Bugfix: properly return errors from the wait in the unary retry backoff. [\#20](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/20) ([devnev](https://github.com/devnev))
- Add 'ErrorToCode' type to support use of custom errors [\#19](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/19) ([adam-26](https://github.com/adam-26))
- Goimports scripts [\#15](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/15) ([mwitkow](https://github.com/mwitkow))
- Generic Logging Middleware [\#14](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/14) ([mwitkow](https://github.com/mwitkow))
- Feature/logging zap [\#13](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/13) ([mwitkow](https://github.com/mwitkow))
- Add array length check to auth metadata [\#12](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/12) ([rodcheater](https://github.com/rodcheater))
- Feature/auth and backoff [\#9](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/9) ([mwitkow](https://github.com/mwitkow))
- Docs/improv [\#8](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/8) ([mwitkow](https://github.com/mwitkow))
- Client Side Retry [\#7](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/7) ([mwitkow](https://github.com/mwitkow))
- Add single key metadata helpers. [\#6](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/6) ([mwitkow](https://github.com/mwitkow))
- Feature/auth middleware [\#5](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/5) ([mwitkow](https://github.com/mwitkow))
- Feature/validator middleware [\#4](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/4) ([mwitkow](https://github.com/mwitkow))
- add streaming client interceptor [\#2](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/2) ([mwitkow](https://github.com/mwitkow))
- Add Chaining for Unary Client Interceptor [\#1](https://github.com/grpc-ecosystem/go-grpc-middleware/pull/1) ([mwitkow](https://github.com/mwitkow))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*