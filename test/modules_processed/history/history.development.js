var Action;
(function(Action2) {
  Action2["Pop"] = "POP";
  Action2["Push"] = "PUSH";
  Action2["Replace"] = "REPLACE";
})(Action || (Action = {}));
const readOnly = (obj) => Object.freeze(obj);
function warning(cond, message) {
  if (!cond) {
    if (typeof console !== "undefined") console.warn(message);
    try {
      throw new Error(message);
    } catch (e) {
    }
  }
}
const BeforeUnloadEventType = "beforeunload";
const HashChangeEventType = "hashchange";
const PopStateEventType = "popstate";
function createBrowserHistory(options = {}) {
  let {
    window = document.defaultView
  } = options;
  let globalHistory = window.history;
  function getIndexAndLocation() {
    let {
      pathname,
      search,
      hash
    } = window.location;
    let state = globalHistory.state || {};
    return [state.idx, readOnly({
      pathname,
      search,
      hash,
      state: state.usr || null,
      key: state.key || "default"
    })];
  }
  let blockedPopTx = null;
  function handlePop() {
    if (blockedPopTx) {
      blockers.call(blockedPopTx);
      blockedPopTx = null;
    } else {
      let nextAction = Action.Pop;
      let [nextIndex, nextLocation] = getIndexAndLocation();
      if (blockers.length) {
        if (nextIndex != null) {
          let delta = index - nextIndex;
          if (delta) {
            blockedPopTx = {
              action: nextAction,
              location: nextLocation,
              retry() {
                go(delta * -1);
              }
            };
            go(delta);
          }
        } else {
          warning(
            false,
            // TODO: Write up a doc that explains our blocking strategy in
            // detail and link to it here so people can understand better what
            // is going on and how to avoid it.
            `You are trying to block a POP navigation to a location that was not created by the history library. The block will fail silently in production, but in general you should do all navigation with the history library (instead of using window.history.pushState directly) to avoid this situation.`
          );
        }
      } else {
        applyTx(nextAction);
      }
    }
  }
  window.addEventListener(PopStateEventType, handlePop);
  let action = Action.Pop;
  let [index, location] = getIndexAndLocation();
  let listeners = createEvents();
  let blockers = createEvents();
  if (index == null) {
    index = 0;
    globalHistory.replaceState(Object.assign(Object.assign({}, globalHistory.state), {
      idx: index
    }), "");
  }
  function createHref(to) {
    return typeof to === "string" ? to : createPath(to);
  }
  function getNextLocation(to, state = null) {
    return readOnly(Object.assign(Object.assign({
      pathname: location.pathname,
      hash: "",
      search: ""
    }, typeof to === "string" ? parsePath(to) : to), {
      state,
      key: createKey()
    }));
  }
  function getHistoryStateAndUrl(nextLocation, index2) {
    return [{
      usr: nextLocation.state,
      key: nextLocation.key,
      idx: index2
    }, createHref(nextLocation)];
  }
  function allowTx(action2, location2, retry) {
    return !blockers.length || (blockers.call({
      action: action2,
      location: location2,
      retry
    }), false);
  }
  function applyTx(nextAction) {
    action = nextAction;
    [index, location] = getIndexAndLocation();
    listeners.call({
      action,
      location
    });
  }
  function push(to, state) {
    let nextAction = Action.Push;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      push(to, state);
    }
    if (allowTx(nextAction, nextLocation, retry)) {
      let [historyState, url] = getHistoryStateAndUrl(nextLocation, index + 1);
      try {
        globalHistory.pushState(historyState, "", url);
      } catch (error) {
        window.location.assign(url);
      }
      applyTx(nextAction);
    }
  }
  function replace(to, state) {
    let nextAction = Action.Replace;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      replace(to, state);
    }
    if (allowTx(nextAction, nextLocation, retry)) {
      let [historyState, url] = getHistoryStateAndUrl(nextLocation, index);
      globalHistory.replaceState(historyState, "", url);
      applyTx(nextAction);
    }
  }
  function go(delta) {
    globalHistory.go(delta);
  }
  let history = {
    get action() {
      return action;
    },
    get location() {
      return location;
    },
    createHref,
    push,
    replace,
    go,
    back() {
      go(-1);
    },
    forward() {
      go(1);
    },
    listen(listener) {
      return listeners.push(listener);
    },
    block(blocker) {
      let unblock = blockers.push(blocker);
      if (blockers.length === 1) {
        window.addEventListener(BeforeUnloadEventType, promptBeforeUnload);
      }
      return function() {
        unblock();
        if (!blockers.length) {
          window.removeEventListener(BeforeUnloadEventType, promptBeforeUnload);
        }
      };
    }
  };
  return history;
}
function createHashHistory(options = {}) {
  let {
    window = document.defaultView
  } = options;
  let globalHistory = window.history;
  function getIndexAndLocation() {
    let {
      pathname = "/",
      search = "",
      hash = ""
    } = parsePath(window.location.hash.substr(1));
    let state = globalHistory.state || {};
    return [state.idx, readOnly({
      pathname,
      search,
      hash,
      state: state.usr || null,
      key: state.key || "default"
    })];
  }
  let blockedPopTx = null;
  function handlePop() {
    if (blockedPopTx) {
      blockers.call(blockedPopTx);
      blockedPopTx = null;
    } else {
      let nextAction = Action.Pop;
      let [nextIndex, nextLocation] = getIndexAndLocation();
      if (blockers.length) {
        if (nextIndex != null) {
          let delta = index - nextIndex;
          if (delta) {
            blockedPopTx = {
              action: nextAction,
              location: nextLocation,
              retry() {
                go(delta * -1);
              }
            };
            go(delta);
          }
        } else {
          warning(
            false,
            // TODO: Write up a doc that explains our blocking strategy in
            // detail and link to it here so people can understand better
            // what is going on and how to avoid it.
            `You are trying to block a POP navigation to a location that was not created by the history library. The block will fail silently in production, but in general you should do all navigation with the history library (instead of using window.history.pushState directly) to avoid this situation.`
          );
        }
      } else {
        applyTx(nextAction);
      }
    }
  }
  window.addEventListener(PopStateEventType, handlePop);
  window.addEventListener(HashChangeEventType, () => {
    let [, nextLocation] = getIndexAndLocation();
    if (createPath(nextLocation) !== createPath(location)) {
      handlePop();
    }
  });
  let action = Action.Pop;
  let [index, location] = getIndexAndLocation();
  let listeners = createEvents();
  let blockers = createEvents();
  if (index == null) {
    index = 0;
    globalHistory.replaceState(Object.assign(Object.assign({}, globalHistory.state), {
      idx: index
    }), "");
  }
  function getBaseHref() {
    let base = document.querySelector("base");
    let href = "";
    if (base && base.getAttribute("href")) {
      let url = window.location.href;
      let hashIndex = url.indexOf("#");
      href = hashIndex === -1 ? url : url.slice(0, hashIndex);
    }
    return href;
  }
  function createHref(to) {
    return getBaseHref() + "#" + (typeof to === "string" ? to : createPath(to));
  }
  function getNextLocation(to, state = null) {
    return readOnly(Object.assign(Object.assign({
      pathname: location.pathname,
      hash: "",
      search: ""
    }, typeof to === "string" ? parsePath(to) : to), {
      state,
      key: createKey()
    }));
  }
  function getHistoryStateAndUrl(nextLocation, index2) {
    return [{
      usr: nextLocation.state,
      key: nextLocation.key,
      idx: index2
    }, createHref(nextLocation)];
  }
  function allowTx(action2, location2, retry) {
    return !blockers.length || (blockers.call({
      action: action2,
      location: location2,
      retry
    }), false);
  }
  function applyTx(nextAction) {
    action = nextAction;
    [index, location] = getIndexAndLocation();
    listeners.call({
      action,
      location
    });
  }
  function push(to, state) {
    let nextAction = Action.Push;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      push(to, state);
    }
    warning(nextLocation.pathname.charAt(0) === "/", `Relative pathnames are not supported in hash history.push(${JSON.stringify(to)})`);
    if (allowTx(nextAction, nextLocation, retry)) {
      let [historyState, url] = getHistoryStateAndUrl(nextLocation, index + 1);
      try {
        globalHistory.pushState(historyState, "", url);
      } catch (error) {
        window.location.assign(url);
      }
      applyTx(nextAction);
    }
  }
  function replace(to, state) {
    let nextAction = Action.Replace;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      replace(to, state);
    }
    warning(nextLocation.pathname.charAt(0) === "/", `Relative pathnames are not supported in hash history.replace(${JSON.stringify(to)})`);
    if (allowTx(nextAction, nextLocation, retry)) {
      let [historyState, url] = getHistoryStateAndUrl(nextLocation, index);
      globalHistory.replaceState(historyState, "", url);
      applyTx(nextAction);
    }
  }
  function go(delta) {
    globalHistory.go(delta);
  }
  let history = {
    get action() {
      return action;
    },
    get location() {
      return location;
    },
    createHref,
    push,
    replace,
    go,
    back() {
      go(-1);
    },
    forward() {
      go(1);
    },
    listen(listener) {
      return listeners.push(listener);
    },
    block(blocker) {
      let unblock = blockers.push(blocker);
      if (blockers.length === 1) {
        window.addEventListener(BeforeUnloadEventType, promptBeforeUnload);
      }
      return function() {
        unblock();
        if (!blockers.length) {
          window.removeEventListener(BeforeUnloadEventType, promptBeforeUnload);
        }
      };
    }
  };
  return history;
}
function createMemoryHistory(options = {}) {
  let {
    initialEntries = ["/"],
    initialIndex
  } = options;
  let entries = initialEntries.map((entry) => {
    let location2 = readOnly(Object.assign({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: createKey()
    }, typeof entry === "string" ? parsePath(entry) : entry));
    warning(location2.pathname.charAt(0) === "/", `Relative pathnames are not supported in createMemoryHistory({ initialEntries }) (invalid entry: ${JSON.stringify(entry)})`);
    return location2;
  });
  let index = clamp(initialIndex == null ? entries.length - 1 : initialIndex, 0, entries.length - 1);
  let action = Action.Pop;
  let location = entries[index];
  let listeners = createEvents();
  let blockers = createEvents();
  function createHref(to) {
    return typeof to === "string" ? to : createPath(to);
  }
  function getNextLocation(to, state = null) {
    return readOnly(Object.assign(Object.assign({
      pathname: location.pathname,
      search: "",
      hash: ""
    }, typeof to === "string" ? parsePath(to) : to), {
      state,
      key: createKey()
    }));
  }
  function allowTx(action2, location2, retry) {
    return !blockers.length || (blockers.call({
      action: action2,
      location: location2,
      retry
    }), false);
  }
  function applyTx(nextAction, nextLocation) {
    action = nextAction;
    location = nextLocation;
    listeners.call({
      action,
      location
    });
  }
  function push(to, state) {
    let nextAction = Action.Push;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      push(to, state);
    }
    warning(location.pathname.charAt(0) === "/", `Relative pathnames are not supported in memory history.push(${JSON.stringify(to)})`);
    if (allowTx(nextAction, nextLocation, retry)) {
      index += 1;
      entries.splice(index, entries.length, nextLocation);
      applyTx(nextAction, nextLocation);
    }
  }
  function replace(to, state) {
    let nextAction = Action.Replace;
    let nextLocation = getNextLocation(to, state);
    function retry() {
      replace(to, state);
    }
    warning(location.pathname.charAt(0) === "/", `Relative pathnames are not supported in memory history.replace(${JSON.stringify(to)})`);
    if (allowTx(nextAction, nextLocation, retry)) {
      entries[index] = nextLocation;
      applyTx(nextAction, nextLocation);
    }
  }
  function go(delta) {
    let nextIndex = clamp(index + delta, 0, entries.length - 1);
    let nextAction = Action.Pop;
    let nextLocation = entries[nextIndex];
    function retry() {
      go(delta);
    }
    if (allowTx(nextAction, nextLocation, retry)) {
      index = nextIndex;
      applyTx(nextAction, nextLocation);
    }
  }
  let history = {
    get index() {
      return index;
    },
    get action() {
      return action;
    },
    get location() {
      return location;
    },
    createHref,
    push,
    replace,
    go,
    back() {
      go(-1);
    },
    forward() {
      go(1);
    },
    listen(listener) {
      return listeners.push(listener);
    },
    block(blocker) {
      return blockers.push(blocker);
    }
  };
  return history;
}
function clamp(n, lowerBound, upperBound) {
  return Math.min(Math.max(n, lowerBound), upperBound);
}
function promptBeforeUnload(event) {
  event.preventDefault();
  event.returnValue = "";
}
function createEvents() {
  let handlers = [];
  return {
    get length() {
      return handlers.length;
    },
    push(fn) {
      handlers.push(fn);
      return function() {
        handlers = handlers.filter((handler) => handler !== fn);
      };
    },
    call(arg) {
      handlers.forEach((fn) => fn && fn(arg));
    }
  };
}
function createKey() {
  return Math.random().toString(36).substr(2, 8);
}
function createPath({
  pathname = "/",
  search = "",
  hash = ""
}) {
  if (search && search !== "?") pathname += search.charAt(0) === "?" ? search : "?" + search;
  if (hash && hash !== "#") pathname += hash.charAt(0) === "#" ? hash : "#" + hash;
  return pathname;
}
function parsePath(path) {
  let parsedPath = {};
  if (path) {
    let hashIndex = path.indexOf("#");
    if (hashIndex >= 0) {
      parsedPath.hash = path.substr(hashIndex);
      path = path.substr(0, hashIndex);
    }
    let searchIndex = path.indexOf("?");
    if (searchIndex >= 0) {
      parsedPath.search = path.substr(searchIndex);
      path = path.substr(0, searchIndex);
    }
    if (path) {
      parsedPath.pathname = path;
    }
  }
  return parsedPath;
}
export { Action, createBrowserHistory, createHashHistory, createMemoryHistory, createPath, parsePath };
