import type { IDateOfBirth } from "../interfaces/IAuth";

export function useValidation() {
  // Validate name
  const validateName = (name: string): string => {
    const nameRegex = /^[a-zA-Z\s]+$/;
    if (!name) {
      return "Name is required";
    } else if (name.length <= 4) {
      return "Name must be more than 4 characters";
    } else if (!nameRegex.test(name)) {
      return "Name must not contain symbols or numbers";
    } else if (name.length > 50) {
      return "Name cannot exceed 50 characters";
    }
    return "";
  };

  // Validate username
  const validateUsername = (username: string): string => {
    if (!username) {
      return "Username is required";
    } else if (username.length < 3) {
      return "Username must be at least 3 characters";
    } else if (username.length > 15) {
      return "Username cannot exceed 15 characters";
    } else if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      return "Username can only contain letters, numbers, and underscores";
    }
    return "";
  };

  // Validate email
  const validateEmail = (email: string): string => {
    const emailRegex = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/;
    if (!email) {
      return "Email is required";
    } else if (!emailRegex.test(email)) {
      return "Please enter a valid email in the format name@domain.com";
    }
    return "";
  };

  // Validate password
  const validatePassword = (password: string): string[] => {
    const errors: string[] = [];

    if (!password) {
      errors.push("Password is required");
      return errors;
    }

    if (password.length < 8) {
      errors.push("Password must be at least 8 characters");
    }

    if (!/[A-Z]/.test(password)) {
      errors.push("Password must contain at least one uppercase letter");
    }

    if (!/[a-z]/.test(password)) {
      errors.push("Password must contain at least one lowercase letter");
    }

    if (!/[0-9]/.test(password)) {
      errors.push("Password must contain at least one number");
    }

    if (!/[^A-Za-z0-9]/.test(password)) {
      errors.push("Password must contain at least one special character");
    }

    return errors;
  };

  // Validate confirm password
  const validateConfirmPassword = (password: string, confirmPassword: string): string => {
    if (!confirmPassword) {
      return "Confirm password is required";
    } else if (password !== confirmPassword) {
      return "Passwords do not match";
    }
    return "";
  };

  // Validate gender
  const validateGender = (gender: string): string => {
    if (!gender) {
      return "Gender is required";
    }
    return "";
  };

  // Validate date of birth
  const validateDateOfBirth = (dateOfBirth: IDateOfBirth, months: string[]): string => {
    if (!dateOfBirth.month || !dateOfBirth.day || !dateOfBirth.year) {
      return "Date of birth is required";
    }

    const birthDate = new Date(
      parseInt(dateOfBirth.year),
      months.indexOf(dateOfBirth.month),
      parseInt(dateOfBirth.day)
    );

    const today = new Date();
    let age = today.getFullYear() - birthDate.getFullYear();
    const monthDiff = today.getMonth() - birthDate.getMonth();

    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
      age--;
    }

    if (age < 13) {
      return "You must be at least 13 years old to register";
    }

    return "";
  };

  // Validate security question
  const validateSecurityQuestion = (question: string, answer: string): string => {
    if (!question) {
      return "Security question is required";
    } else if (!answer) {
      return "Security answer is required";
    } else if (answer.length < 3) {
      return "Security answer must be at least 3 characters";
    }
    return "";
  };

  // Format date of birth for API
  const formatDateOfBirth = (dateOfBirth: IDateOfBirth, months: string[]): string => {
    const month = (months.indexOf(dateOfBirth.month) + 1).toString().padStart(2, "0");
    const day = dateOfBirth.day.padStart(2, "0");
    return `${dateOfBirth.year}-${month}-${day}`;
  };

  return {
    validateName,
    validateUsername,
    validateEmail,
    validatePassword,
    validateConfirmPassword,
    validateGender,
    validateDateOfBirth,
    validateSecurityQuestion,
    formatDateOfBirth
  };
}