chrome.app.runtime.onLaunched.addListener(function() {
  chrome.app.window.create('viewer.html', {
    id: "gofer",
    bounds: {
      width: 800,
      height: 600
    }
  });
});
