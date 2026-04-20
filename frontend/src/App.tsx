import { useState, useEffect, useCallback } from "react";
import type { TodoList, TodoItem } from "./gen/showcase/app/todolist/v1/todolist_v1_pb.js";
import { State } from "./gen/showcase/app/todolist/v1/todolist_v1_pb.js";
import type { History } from "@protosource/client";
import { APIError } from "@protosource/client";
import { create as createProto } from "@bufbuild/protobuf";
import { TodoItemSchema } from "./gen/showcase/app/todolist/v1/todolist_v1_pb.js";
import type { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";
import "./App.css";

function generateId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    return (c === "x" ? r : (r & 0x3) | 0x8).toString(16);
  });
}

interface AppProps {
  client: TodoListHTTPClient;
  actor: string;
}

export default function App({ client: todoListClient, actor }: AppProps) {
  const [allLists, setAllLists] = useState<TodoList[]>([]);
  const [showArchived, setShowArchived] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [list, setList] = useState<TodoList | null>(null);
  const [history, setHistory] = useState<History | null>(null);
  const [showHistory, setShowHistory] = useState(false);
  const [newListName, setNewListName] = useState("");
  const [newItemTitle, setNewItemTitle] = useState("");
  const [renamingTo, setRenamingTo] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadMyLists = useCallback(async () => {
    try {
      let results: TodoList[];
      if (showArchived) {
        results = await todoListClient.queryByCreateBy(actor);
      } else {
        results = await todoListClient.queryByCreateByWithState(
          actor, "eq", State.ACTIVE
        );
      }
      setAllLists(results);
      setError(null);
    } catch (e: unknown) {
      if (e instanceof APIError && e.statusCode === 404) {
        setAllLists([]);
        setError(null);
      } else {
        setError(`Failed to load lists: ${e instanceof Error ? e.message : e}`);
      }
    }
  }, [showArchived, todoListClient, actor]);

  useEffect(() => {
    loadMyLists();
  }, [loadMyLists]);

  const loadList = useCallback(async (id: string) => {
    try {
      const loaded = await todoListClient.load(id);
      setList(loaded);
      setAllLists((prev) =>
        prev.map((entry) => (entry.id === id ? loaded : entry))
      );
      setError(null);
    } catch (e: unknown) {
      setError(`Failed to load list: ${e instanceof Error ? e.message : e}`);
    }
  }, [todoListClient]);

  const loadHistory = useCallback(async (id: string) => {
    try {
      const h = await todoListClient.history(id);
      setHistory(h);
    } catch (e: unknown) {
      setError(`Failed to load history: ${e instanceof Error ? e.message : e}`);
    }
  }, [todoListClient]);

  useEffect(() => {
    if (selectedId) {
      loadList(selectedId);
      if (showHistory) loadHistory(selectedId);
    }
  }, [selectedId, showHistory, loadList, loadHistory]);

  async function handleCreateList(e: React.FormEvent) {
    e.preventDefault();
    if (!newListName.trim()) return;
    const id = generateId();
    const name = newListName.trim();
    try {
      await todoListClient.create(id, name);
      setNewListName("");
      setSelectedId(id);
      await loadMyLists();
      await loadList(id);
      setError(null);
    } catch (err: unknown) {
      setError(`Create failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleRename() {
    if (!selectedId || !renamingTo?.trim()) return;
    try {
      await todoListClient.rename(selectedId, renamingTo.trim());
      setRenamingTo(null);
      await loadList(selectedId);
    } catch (err: unknown) {
      setError(`Rename failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleArchive() {
    if (!selectedId) return;
    try {
      await todoListClient.archive(selectedId);
      await loadList(selectedId);
      await loadMyLists();
    } catch (err: unknown) {
      setError(`Archive failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleUnarchive() {
    if (!selectedId) return;
    try {
      await todoListClient.unarchive(selectedId);
      await loadList(selectedId);
      await loadMyLists();
    } catch (err: unknown) {
      setError(`Unarchive failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleAddItem(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedId || !newItemTitle.trim()) return;
    const item = createProto(TodoItemSchema, {
      itemId: generateId(),
      title: newItemTitle.trim(),
      completed: false,
      position: list ? list.itemCount + 1 : 1,
      createdAt: BigInt(Date.now()),
    });
    try {
      await todoListClient.addItem(selectedId, item);
      setNewItemTitle("");
      await loadList(selectedId);
    } catch (err: unknown) {
      setError(`Add item failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleToggleItem(item: TodoItem) {
    if (!selectedId) return;
    const updated = createProto(TodoItemSchema, {
      itemId: item.itemId,
      title: item.title,
      completed: !item.completed,
      position: item.position,
      createdAt: item.createdAt,
    });
    try {
      await todoListClient.updateItem(selectedId, updated);
      await loadList(selectedId);
    } catch (err: unknown) {
      setError(`Update failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  async function handleRemoveItem(itemId: string) {
    if (!selectedId) return;
    try {
      await todoListClient.removeItem(selectedId, itemId);
      await loadList(selectedId);
    } catch (err: unknown) {
      setError(`Remove failed: ${err instanceof Error ? err.message : err}`);
    }
  }

  const items = list ? Object.values(list.items) : [];
  const isArchived = list?.state === 2; // STATE_ARCHIVED

  return (
    <div className="app">
      <header>
        <h1>todoapp</h1>
        <p className="subtitle">protosource showcase</p>
      </header>

      {error && (
        <div className="error" onClick={() => setError(null)}>
          {error}
        </div>
      )}

      <div className="layout">
        <aside className="sidebar">
          <form onSubmit={handleCreateList} className="create-form">
            <input
              type="text"
              value={newListName}
              onChange={(e) => setNewListName(e.target.value)}
              placeholder="New list name..."
            />
            <button type="submit">Create</button>
          </form>

          <label className="show-archived">
            <input
              type="checkbox"
              checked={showArchived}
              onChange={(e) => setShowArchived(e.target.checked)}
            />
            Show archived
          </label>

          <ul className="list-selector">
            {allLists.map((entry) => (
              <li
                key={entry.id}
                className={`${entry.id === selectedId ? "selected" : ""} ${entry.state === 2 ? "archived-entry" : ""}`}
                onClick={() => {
                  setSelectedId(entry.id);
                  setShowHistory(false);
                }}
              >
                <span className="list-name">{entry.name}</span>
              </li>
            ))}
            {allLists.length === 0 && (
              <li className="empty-items">No lists yet.</li>
            )}
          </ul>
        </aside>

        <main className="content">
          {!selectedId && (
            <div className="empty-state">
              Create or select a list to get started.
            </div>
          )}

          {selectedId && list && (
            <>
              <div className="list-header">
                {renamingTo !== null ? (
                  <div className="rename-form">
                    <input
                      autoFocus
                      value={renamingTo}
                      onChange={(e) => setRenamingTo(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleRename();
                        if (e.key === "Escape") setRenamingTo(null);
                      }}
                    />
                    <button onClick={handleRename}>Save</button>
                    <button onClick={() => setRenamingTo(null)}>Cancel</button>
                  </div>
                ) : (
                  <h2 onClick={() => setRenamingTo(list.name)}>{list.name}</h2>
                )}

                <div className="list-meta">
                  <span className={`state ${isArchived ? "archived" : "active"}`}>
                    {isArchived ? "Archived" : "Active"}
                  </span>
                  <span className="counts">
                    {list.completedCount}/{list.itemCount} done
                  </span>
                  <span className="version">v{list.version.toString()}</span>
                </div>

                <div className="list-actions">
                  {isArchived ? (
                    <button onClick={handleUnarchive}>Unarchive</button>
                  ) : (
                    <button onClick={handleArchive}>Archive</button>
                  )}
                  <button
                    className={showHistory ? "active" : ""}
                    onClick={() => {
                      setShowHistory(!showHistory);
                      if (!showHistory) loadHistory(selectedId);
                    }}
                  >
                    History
                  </button>
                </div>
              </div>

              {!isArchived && (
                <form onSubmit={handleAddItem} className="add-item-form">
                  <input
                    type="text"
                    value={newItemTitle}
                    onChange={(e) => setNewItemTitle(e.target.value)}
                    placeholder="Add an item..."
                  />
                  <button type="submit">Add</button>
                </form>
              )}

              <ul className="items">
                {items.map((item) => (
                  <li key={item.itemId} className={item.completed ? "completed" : ""}>
                    <label>
                      <input
                        type="checkbox"
                        checked={item.completed}
                        onChange={() => handleToggleItem(item)}
                        disabled={isArchived}
                      />
                      <span className="item-title">{item.title}</span>
                    </label>
                    {!isArchived && (
                      <button
                        className="btn-remove"
                        onClick={() => handleRemoveItem(item.itemId)}
                      >
                        x
                      </button>
                    )}
                  </li>
                ))}
                {items.length === 0 && (
                  <li className="empty-items">No items yet.</li>
                )}
              </ul>

              {showHistory && history && (
                <div className="history-panel">
                  <h3>Event History ({history.records.length} events)</h3>
                  <table>
                    <thead>
                      <tr>
                        <th>Version</th>
                        <th>Size</th>
                      </tr>
                    </thead>
                    <tbody>
                      {history.records.map((rec) => (
                        <tr key={rec.version.toString()}>
                          <td>{rec.version.toString()}</td>
                          <td>{rec.data.length} bytes</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          )}
        </main>
      </div>
    </div>
  );
}
