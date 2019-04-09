import { h, app } from "hyperapp";

const state = {
  docs: [{title: "test"}],
  currentDoc: {
    title: "",
    path: "",
    content: "",
  },
};

const actions = {
  setTitle: title => (state, actions) => {
    let newState = {
      currentDoc: {
        ...state.currentDoc,
        title: title,
      }
    };

    if (!state.path) {
      const words = title.toLowerCase().split(' ');
      newState.currentDoc.path = words.join('-');
      console.log(newState);
    }

    return newState;
  },
  setPath: path => state => ({
    currentDoc: {
      ...state.currentDoc,
      path: path,
    }
  }),
  setContent: content => (state, actions) => {
    return {
      currentDoc: {
        ...state.currentDoc,
        content: content,
      }
    };
  },
  setDocs: value => state => {
    console.log("Received:", value);
    return {docs: value}
  },
  refreshDocs: () => (state, actions) => {
    return fetch("http://localhost:1337/api/doc/list")
      .then(res => res.json())
      .then(actions.setDocs);
  },
  submitDoc: value => (state, actions) => {
    return post('http://localhost:1337/api/doc', state.currentDoc)
      .then(actions.refreshDocs())
      .catch(console.log);
  },
  clear: value => state => {
    return state;
  }
};

function view(state, actions) {
  const { currentDoc } = state;
  return (
    <div>
      <h2>DocShelf PoC</h2>
      <ul>
        { state.docs.map(doc => <li>{doc.title}</li>) }
      </ul>
      <button onclick={actions.refreshDocs}>Refresh</button>

      <div>
        <input oninput={handleSetTitle(actions)} value={currentDoc.title}/>
      </div>
      <div>
        <input oninput={handleSetPath(actions)} value={currentDoc.path}/>
      </div>
      <div>
        <textarea oninput={handleSetContent(actions)} value={currentDoc.content}></textarea>
      </div>
      <div>
        <button onclick={actions.submitDoc}>Submit Doc</button>
      </div>
    </div>
  );
}

app(state, actions, view, document.body);

// define functions outside of view so we don't redeclare them every time
function handleSetTitle(actions) {
  return function(event) {
    return actions.setTitle(event.target.value);
  }
}

function handleSetPath(actions) {
  return function(event) {
    return actions.setPath(event.target.value);
  }
}

function handleSetContent(actions) {
  return function(event) {
    return actions.setContent(event.target.value);
  }
}

function submitDoc() {
  return function() {
    return function(state, actions) {
    };
  }
}

function post(url = ``, data = {}) {
  // Default options are marked with *
    return fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
    })
    .then(res => {
      if (res.status === 200) {
        return res.json();
      }
    });
}
