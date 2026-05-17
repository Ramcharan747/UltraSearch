package solver

// StealthScript is a comprehensive JS payload that patches all known bot detection signals.
// This must be injected via page.AddScriptToEvaluateOnNewDocument BEFORE any navigation.
const StealthScript = `
// 1. Remove webdriver flag
Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
delete navigator.__proto__.webdriver;

// 2. Mock chrome.runtime to look like a real browser extension context
if (!window.chrome) window.chrome = {};
if (!window.chrome.runtime) {
  window.chrome.runtime = {
    connect: function() {},
    sendMessage: function() {},
    onMessage: { addListener: function() {} },
    id: undefined
  };
}

// 3. Override navigator.plugins to return realistic plugin list
Object.defineProperty(navigator, 'plugins', {
  get: () => {
    const fakePlugins = [
      { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format', length: 1 },
      { name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: '', length: 1 },
      { name: 'Native Client', filename: 'internal-nacl-plugin', description: '', length: 2 },
    ];
    fakePlugins.refresh = () => {};
    fakePlugins.item = (i) => fakePlugins[i];
    fakePlugins.namedItem = (name) => fakePlugins.find(p => p.name === name);
    Object.defineProperty(fakePlugins, 'length', { get: () => 3 });
    return fakePlugins;
  }
});

// 4. Override navigator.languages
Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });

// 5. Fix permissions query to not leak automation
const originalQuery = window.Permissions && window.Permissions.prototype.query;
if (originalQuery) {
  window.Permissions.prototype.query = function(parameters) {
    if (parameters.name === 'notifications') {
      return Promise.resolve({ state: Notification.permission });
    }
    return originalQuery.call(this, parameters);
  };
}

// 6. Prevent iframe contentWindow detection
try {
  const elementDescriptor = Object.getOwnPropertyDescriptor(HTMLIFrameElement.prototype, 'contentWindow');
  if (elementDescriptor) {
    Object.defineProperty(HTMLIFrameElement.prototype, 'contentWindow', {
      get: function() {
        return elementDescriptor.get.apply(this);
      }
    });
  }
} catch (e) {}

// 7. Fix WebGL vendor/renderer to not leak headless
const getParameter = WebGLRenderingContext.prototype.getParameter;
WebGLRenderingContext.prototype.getParameter = function(parameter) {
  if (parameter === 37445) return 'Intel Inc.';
  if (parameter === 37446) return 'Intel Iris OpenGL Engine';
  return getParameter.call(this, parameter);
};

// 8. Fix toString of patched functions so they pass Function.prototype.toString checks
const nativeToStringFunctionString = Error.toString().replace(/Error/g, "toString");
const oldToString = Function.prototype.toString;
function hookedToString() {
  if (this === window.navigator.webdriver) return "function webdriver() { [native code] }";
  if (this === hookedToString) return nativeToStringFunctionString;
  return oldToString.call(this);
}
Function.prototype.toString = hookedToString;

// 9. Remove CDP artifacts from window
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;

// 10. Fake window.outerHeight/outerWidth to not leak headless
if (window.outerHeight === 0) {
  Object.defineProperty(window, 'outerHeight', { get: () => window.innerHeight + 85 });
}
if (window.outerWidth === 0) {
  Object.defineProperty(window, 'outerWidth', { get: () => window.innerWidth });
}

// 11. Override navigator.hardwareConcurrency
Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => 8 });

// 12. Override navigator.deviceMemory
Object.defineProperty(navigator, 'deviceMemory', { get: () => 8 });

// 13. Fix connection property
if (navigator.connection) {
  Object.defineProperty(navigator.connection, 'rtt', { get: () => 50 });
}
`

// StealthAllocatorFlags returns the Chrome flags needed for maximum stealth
func StealthAllocatorFlags() map[string]interface{} {
	return map[string]interface{}{
		"disable-blink-features":  "AutomationControlled",
		"disable-infobars":        true,
		"enable-automation":       false,
		"disable-background-networking":                false,
		"disable-component-update":                     false,
		"disable-default-apps":                         false,
		"disable-extensions":                           false,
		"excludeSwitches":                               "enable-automation",
	}
}
