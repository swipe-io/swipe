# Swipe JSONRPC Client

<a href="#Create">Create</a>

<a href="#Delete">Delete</a>

<a href="#Get">Get</a>

<a href="#GetAll">GetAll</a>

<a href="#TestMethod">TestMethod</a>

<a href="#TestMethod2">TestMethod2</a>

### <a name="Create"></a> Create(name, data) ⇒<code>void</code>

 new item of item.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|name|<code>string</code>||
|data|<code>Array.&lt;number&gt;</code>||
### <a name="Delete"></a> Delete(id) ⇒



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
### <a name="Get"></a> Get(id, name, fname, price, n, b, c) ⇒<code><a href="#User">User</a></code>

 item.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
|name|<code>string</code>||
|fname|<code>string</code>||
|price|<code>number</code>||
|n|<code>number</code>||
|b|<code>number</code>||
|c|<code>number</code>||
### <a name="GetAll"></a> GetAll() ⇒<code>Array.&lt;<a href="#User">User</a>&gt;</code>

 more comment and more and more comment and more and more comment and more.

New line comment.



**Throws**:

<code>ErrUnauthorizedException</code>



### <a name="TestMethod"></a> TestMethod(data, ss) ⇒<code>Object.&lt;string, Object.&lt;string, Array.&lt;string&gt;&gt;&gt;</code>



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|data|<code>Object.&lt;string, Object&gt;</code>||
|ss|<code>Object</code>||
### <a name="TestMethod2"></a> TestMethod2(ns, utype, user, restype, resource, permission) ⇒<code>void</code>



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|ns|<code>string</code>||
|utype|<code>string</code>||
|user|<code>string</code>||
|restype|<code>string</code>||
|resource|<code>string</code>||
|permission|<code>string</code>||
## Members

### User

| Field | Type | Description |
|------|------|------|
|id|<code>string</code>||
|name|<code>string</code>||
|password|<code>string</code>||
|point|<code><a href="#GeoJSON">GeoJSON</a></code>||
|last_seen|<code>string</code>||
|photo|<code>Array.&lt;number&gt;</code>||
|profile|<code><a href="#Profile">Profile</a></code>||
|created_at|<code>string</code>||
|updated_at|<code>string</code>||
### GeoJSON

| Field | Type | Description |
|------|------|------|
|coordinates200|<code>Array.&lt;number&gt;</code>||
### Profile

| Field | Type | Description |
|------|------|------|
|phone|<code>string</code>||
