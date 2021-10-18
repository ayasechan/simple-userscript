# simple-userscript

Automatically update web pages when user scripts change

Suitable for simple development of **single-file** scripts

# usage

Assuming there is an `example.js` file in the working directory
The content of the file is as follows

```js
// ==UserScript==
// @name         set title as hello
// @version      0.0.1
// @description  my awesome script
// @match        *://*/*
// @run-at document-end
// ==/UserScript==

(() => {
  const main = () => {
    document.title = 'hello';
  };

  main();

})();
```
We can use `simple-userscript -f example.js` to start a service

Follow the prompts to install the development script

Then open any web page and modify `example.js`

You will see the webpage refreshes automatically


