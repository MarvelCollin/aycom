FROM node:18

WORKDIR /app

COPY package*.json ./

RUN npm install

# Install sass-embedded dependency
RUN npm install -D sass-embedded

COPY . .

RUN npm run build

# Install serve to run the application
RUN npm install -g serve

EXPOSE 3000

CMD ["serve", "-s", "dist", "-l", "3000"] 