export const getUserMe = async () => {
  const response = await fetch("/api/v1/users/me", {
    credentials: 'same-origin'
  });
  if (!response.ok) throw new Error("Not logged in");
  return response.json(); // Returns { id, username, role }
};