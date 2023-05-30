# A tool for static website deployment 

Build the static html website in current directory into `dist/` folder, and open a `Github Pages`-like web server to see the result.

```shell
static --open
```

# Embeding other html

We support html embeding

`index.html`
```html
<html>
    <body>
        {{template "component/header.html"}}
        <span>Home</span>
    </body>
</html>
```

`component/header.html`
```html
<h1>Header</h1>
```
