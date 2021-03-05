# UsersvitalyworkgosrcgithubComswipeIoswipev2fixturesServiceJSONRPCMultiapp JSONRPC Client

## Getting Started

You can install this with:

```shell script
npm install --save-dev service
```

Import the package with the client:

```javascript
import API from "service"
```

Create a transport, only one method needs to be implemented: `doRequest(Array.<Object>) PromiseLike<Object>`.

For example:

```javascript
class FetchTransport {
    constructor(url) {
      this.url = url;
    }

    doRequest(requests) {
        return fetch(this.url, {method: "POST", body: JSON.stringify(requests)})
    }
}
```

Now for a complete example:

```javascript
import API from "service"
import Transport from "transport"

const api = new API(new Transport("http://127.0.0.1"))

// call method here.
```

## API
## Methods

<a href="#a.TestMethod">a.TestMethod</a>

### <a name="a.TestMethod"></a>a.TestMethod() ⇒<code>void</code>





**Throws**:



<a href="#b.Create">b.Create</a>

<a href="#b.Delete">b.Delete</a>

<a href="#b.Get">b.Get</a>

<a href="#b.GetAll">b.GetAll</a>

<a href="#b.TestMethod">b.TestMethod</a>

<a href="#b.TestMethod2">b.TestMethod2</a>

### <a name="b.Create"></a>b.Create(newData, name, data) ⇒<code>void</code>

 new item of item.



**Throws**:

<code>ErrUnauthorizedException</code>



| Param | Type | Description |
|------|------|------|
|newData|<code><a href="#Data">Data</a></code>||
|name|<code>string</code>||
|data|<code>Array.&lt;number&gt;</code>||
### <a name="b.Delete"></a>b.Delete(id) ⇒





**Throws**:



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
### <a name="b.Get"></a>b.Get(id, name, fname, price, n, b, cc) ⇒<code><a href="#User">User</a></code>

 item.



**Throws**:



| Param | Type | Description |
|------|------|------|
|id|<code>number</code>||
|name|<code>string</code>||
|fname|<code>string</code>||
|price|<code>number</code>||
|n|<code>number</code>||
|b|<code>number</code>||
|cc|<code>number</code>||
### <a name="b.GetAll"></a>b.GetAll(members) ⇒<code>Array.&lt;<a href="#User">User</a>&gt;</code>

 more comment and more and more comment and more and more comment and more.New line comment.



**Throws**:



| Param | Type | Description |
|------|------|------|
|members|<code><a href="#Members">Members</a></code>||
### <a name="b.TestMethod"></a>b.TestMethod(data, ss) ⇒<code>Object.&lt;string, Object.&lt;string, Array.&lt;string&gt;&gt;&gt;</code>





**Throws**:



| Param | Type | Description |
|------|------|------|
|data|<code>Object.&lt;string, Object&gt;</code>||
|ss|<code>Object</code>||
### <a name="b.TestMethod2"></a>b.TestMethod2(ns, utype, user, restype, resource, permission) ⇒<code>void</code>





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
|title|<code>string</code>||
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
