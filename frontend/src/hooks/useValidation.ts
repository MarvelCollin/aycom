import type { DateOfBirth } from '../interfaces/auth';

export function useValidation() {
  // Validate name
  const validateName = (name: string): string => {
    if (name.length < 4) {
      return "Name must be at least 4 characters long";
    }
    
    if (/[0-9!@#$%^&*(),.?":{}|<>]/.test(name)) {
      return "Name cannot contain numbers or symbols";
    }
    
    return "";
  };
  
  // Validate username
  const validateUsername = (username: string): string => {
    if (username.length < 4) {
      return "Username must be at least 4 characters long";
    }
    
    // Ini mau check dari db lagi
    return "";
  };
  
  // Validate email
  const validateEmail = (email: string): string => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return "Please enter a valid email address (e.g., name@domain.com)";
    }
    
    // In a real app, we would check if the email is unique via API call
    return "";
  };
  
  // Validate password
  const validatePassword = (password: string): string[] => {
    const errors: string[] = [];
    
    if (password.length < 8) {
      errors.push("Password must be at least 8 characters long");
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
    
    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      errors.push("Password must contain at least one special character");
    }
    
    return errors;
  };
  
  // Validate confirm password
  const validateConfirmPassword = (password: string, confirmPassword: string): string => {
    if (password !== confirmPassword) {
      return "Passwords do not match";
    }
    
    return "";
  };
  
  // Validate gender
  const validateGender = (gender: string): string => {
    if (!gender) {
      return "Please select your gender";
    }
    
    return "";
  };
  
  // Validate date of birth
  const validateDateOfBirth = (dateOfBirth: DateOfBirth, months: string[]): string => {
    if (!dateOfBirth.month || !dateOfBirth.day || !dateOfBirth.year) {
      return "Please select your date of birth";
    }
    
    const dob = new Date(
      parseInt(dateOfBirth.year),
      months.indexOf(dateOfBirth.month),
      parseInt(dateOfBirth.day)
    );
    
    const today = new Date();
    const age = today.getFullYear() - dob.getFullYear();
    const m = today.getMonth() - dob.getMonth();
    
    // If birth month is after current month or
    // birth month is current month but birth day is after current day
    if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) {
      if (age - 1 < 13) {
        return "You must be at least 13 years old";
      }
    } else {
      if (age < 13) {
        return "You must be at least 13 years old";
      }
    }
    
    return "";
  };
  
  // Validate security question
  const validateSecurityQuestion = (question: string, answer: string): string => {
    if (!question || !answer) {
      return "Please select a security question and provide an answer";
    }
    
    return "";
  };
  
  // Format date of birth for API
  const formatDateOfBirth = (dateOfBirth: DateOfBirth, months: string[]): string => {
    if (!dateOfBirth.month || !dateOfBirth.day || !dateOfBirth.year) {
      return "";
    }
    
    const month = months.indexOf(dateOfBirth.month) + 1;
    const paddedMonth = month < 10 ? `0${month}` : month;
    const paddedDay = parseInt(dateOfBirth.day) < 10 ? `0${dateOfBirth.day}` : dateOfBirth.day;
    
    return `${dateOfBirth.year}-${paddedMonth}-${paddedDay}`;
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