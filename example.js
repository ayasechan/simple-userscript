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