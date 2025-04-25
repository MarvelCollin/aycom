FROM node:18

WORKDIR /app

COPY package*.json ./

RUN npm install

RUN npm install -D sass-embedded

COPY . .

RUN npm run build

RUN npm install -g serve

EXPOSE 3000

CMD ["serve", "-s", "dist", "-l", "3000"] 