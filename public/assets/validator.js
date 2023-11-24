const passwordValidator = (text) => {
  if (text?.length === 0) return null;
  if (text?.includes(" ")) return "space not allowed";
  return !RegExp(/^[a-zA-Z0-9!@#$&'*+?^_-]+$/).test(String(text))
    ? "Allowed characters: !@#$&'*+?^_-"
    : (text?.length || 0) < 6
    ? "password too short"
    : null;
};
