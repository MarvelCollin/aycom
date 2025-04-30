export async function createThread(data: Record<string, any>) {
  const response = await fetch("/api/threads", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to create thread");
  return response.json();
}

export async function getThread(id: string) {
  const response = await fetch(`/api/threads/${id}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to fetch thread");
  return response.json();
}

export async function getThreadsByUser(userId: string) {
  const response = await fetch(`/api/threads/user/${userId}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to fetch user's threads");
  return response.json();
}

export async function updateThread(id: string, data: Record<string, any>) {
  const response = await fetch(`/api/threads/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to update thread");
  return response.json();
}

export async function deleteThread(id: string) {
  const response = await fetch(`/api/threads/${id}`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to delete thread");
  return response.json();
} 