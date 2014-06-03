chrome.app.runtime.onLaunched.addListener(function() {
  chrome.app.window.create('viewer.html', {
    outerBounds: {
      width: 800,
      height: 600
    }
  });
});
