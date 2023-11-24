document.addEventListener("alpine:init", () => {
  window.Alpine.data("state", () => ({
    password: "",
    confirmPassword: "",
    message: null,
    loading: false,

    get isValid() {
      return (
        this.password.length > 0 &&
        passwordValidator(this.password) == null &&
        this.password === this.confirmPassword
      );
    },

    async save(token) {
      this.loading = true;
      this.message = null;
      try {
        const response = await fetch("/password-reset", {
          method: "PUT",
          body: JSON.stringify({
            password: this.password,
            token,
          }),
        });

        switch (response.status) {
          case 400:
            this.message = "validation failed";
            break;
          case 200:
            let res = await response.json();
            this.message = res.message;
            return;
          case 202:
            document.getElementById("app").innerHTML = await response.text();
        }
      } catch (e) {
        this.message = "Something went wrong";
      }
      this.loading = false;
    },

    init() {
      this.$watch("password", (v) => {
        if (passwordValidator(this.password)) {
          return (this.message = passwordValidator(this.password));
        }
        if (this.password !== this.confirmPassword) {
          return (this.message = "password does not match");
        }
        this.message = "";
      });
      this.$watch("confirmPassword", (v) => {
        if (this.password !== this.confirmPassword) {
          return (this.message = "password does not match");
        }
        this.message = "";
      });
    },
  }));
});
