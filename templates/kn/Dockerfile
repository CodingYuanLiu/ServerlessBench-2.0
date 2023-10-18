FROM python

WORKDIR /app

RUN pip install flask

COPY . .

RUN pip install --no-cache-dir -r requirements.txt

ENV FLASK_APP=server

EXPOSE 8080

CMD ["python", "server.py"]
