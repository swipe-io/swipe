# Swipe JSONRPC Client

<a href="#interfaceA.TestMethod">interfaceA.TestMethod</a>

### <a name="interfaceA.TestMethod"></a>interfaceA.TestMethod() ⇒<code>void</code>



**Throws**:

<code>ErrUnauthorizedException</code>



<a href="#interfaceB.Create">interfaceB.Create</a>

<a href="#interfaceB.Delete">interfaceB.Delete</a>

<a href="#interfaceB.Get">interfaceB.Get</a>

<a href="#interfaceB.GetAll">interfaceB.GetAll</a>

<a href="#interfaceB.TestMethod">interfaceB.TestMethod</a>

<a href="#interfaceB.TestMethod2">interfaceB.TestMethod2</a>

### <a name="interfaceB.Create"></a>interfaceB.Create(newData, name, data) ⇒<code>void</code>

 new item of item.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|newData|<code><a href="#Data">Data</a></code>||
|name|<code>string</code>||
|data|<code>Array.&lt;number&gt;</code>||
### <a name="interfaceB.Delete"></a>interfaceB.Delete(id) ⇒



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
### <a name="interfaceB.Get"></a>interfaceB.Get(id, name, fname, price, n, b, cc) ⇒<code><a href="#User">User</a></code>

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
|cc|<code>number</code>||
### <a name="interfaceB.GetAll"></a>interfaceB.GetAll(members) ⇒<code>Array.&lt;<a href="#User">User</a>&gt;</code>

 more comment and more and more comment and more and more comment and more.

New line comment.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|members|<code><a href="#Members">Members</a></code>||
### <a name="interfaceB.TestMethod"></a>interfaceB.TestMethod(data, ss) ⇒<code>Object.&lt;string, Object.&lt;string, Array.&lt;string&gt;&gt;&gt;</code>



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|data|<code>Object.&lt;string, Object&gt;</code>||
|ss|<code>Object</code>||
### <a name="interfaceB.TestMethod2"></a>interfaceB.TestMethod2(ns, utype, user, restype, resource, permission) ⇒<code>void</code>



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
|created_at|<code>string</code>||
|updated_at|<code>string</code>||
