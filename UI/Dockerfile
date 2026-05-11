FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build -- --configuration production

FROM nginx:stable-alpine
RUN rm -rf /usr/share/nginx/html/*
# Копируем собранную статику из этапа сборки
COPY --from=build /app/dist/web /usr/share/nginx/html
# Выставляем стандартный порт Nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]