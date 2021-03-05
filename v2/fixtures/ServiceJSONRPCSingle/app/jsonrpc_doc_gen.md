# UsersvitalyworkgosrcgithubComswipeIoswipev2fixturesServiceJSONRPCSingleapp JSONRPC Client

<a href="#Create">Create</a>

<a href="#Delete">Delete</a>

<a href="#Get">Get</a>

<a href="#GetAll">GetAll</a>

<a href="#TestMethod">TestMethod</a>

<a href="#TestMethod2">TestMethod2</a>

### <a name="Create"></a>Create(newData, name, data) ⇒<code>void</code>

 new item of item.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|newData|<code><a href="#Data">Data</a></code>||
|name|<code>string</code>||
|data|<code>Array.&lt;number&gt;</code>||
### <a name="Delete"></a>Delete(id) ⇒





**Throws**:



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
### <a name="Get"></a>Get(id, name, fname, price, n, b, cc) ⇒<code><a href="#User">User</a></code>

 item.



**Throws**:



| Param | Type | Description |
|------|------|------|
|id|<code>string</code>||
|name|<code>string</code>||
|fname|<code>string</code>||
|price|<code>number</code>||
|n|<code>number</code>||
|b|<code>number</code>||
|cc|<code>number</code>||
### <a name="GetAll"></a>GetAll(members) ⇒<code>Array.&lt;<a href="#User">User</a>&gt;</code>

 more comment and more and more comment and more and more comment and more.New line comment.



**Throws**:



| Param | Type | Description |
|------|------|------|
|members|<code><a href="#Members">Members</a></code>||
### <a name="TestMethod"></a>TestMethod(data, ss) ⇒<code>Object.&lt;string, Object.&lt;string, Array.&lt;string&gt;&gt;&gt;</code>





**Throws**:



| Param | Type | Description |
|------|------|------|
|data|<code>Object.&lt;string, Object&gt;</code>||
|ss|<code>Object</code>||
### <a name="TestMethod2"></a>TestMethod2(ns, utype, user, restype, resource, permission) ⇒<code>void</code>





**Throws**:



| Param | Type | Description |
|------|------|------|
|ns|<code>string</code>||
|utype|<code>string</code>||
|user|<code>string</code>||
|restype|<code>string</code>||
|resource|<code>string</code>||
|permission|<code>string</code>||
## Members

### GeoJSON

| Field | Type | Description |
|------|------|------|
|coordinates200|<code>Array.&lt;number&gt;</code>||
### Profile

| Field | Type | Description |
|------|------|------|
|phone|<code>string</code>||
### Recurse

| Field | Type | Description |
|------|------|------|
|name|<code>string</code>||
|recurse|<code>Array.&lt;<a href="#Recurse">Recurse</a>&gt;</code>||
### User

| Field | Type | Description |
|------|------|------|
|id|<code>string</code>||
|name|<code>string</code>||
|password|<code>string</code>||
|point|<code><a href="#GeoJSON">GeoJSON</a></code>||
|last_seen|<code>string</code>||
|data|<code><a href="#Data">Data</a></code>||
|photo|<code>Array.&lt;number&gt;</code>||
|user|<code><a href="#User">User</a></code>||
|profile|<code><a href="#Profile">Profile</a></code>||
|recurse|<code><a href="#Recurse">Recurse</a></code>||
|kind|<code><a href="#Kind">Kind</a></code>||
|created_at|<code>string</code>||
|updated_at|<code>string</code>||
